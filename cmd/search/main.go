package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ericvolp12/bsky-experiments/pkg/graph"
	"github.com/ericvolp12/bsky-experiments/pkg/layout"
	"github.com/ericvolp12/bsky-experiments/pkg/search"
	ginprometheus "github.com/ericvolp12/go-gin-prometheus"
	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

// Initialize Prometheus Metrics for cache hits and misses
var cacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "bsky_cache_hits_total",
	Help: "The total number of cache hits",
}, []string{"cache_type"})

var cacheMisses = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "bsky_cache_misses_total",
	Help: "The total number of cache misses",
}, []string{"cache_type"})

type ThreadViewCacheEntry struct {
	ThreadView []search.PostView
	Expiration time.Time
}

type LayoutCacheEntry struct {
	Layout     []layout.ThreadViewLayout
	Expiration time.Time
}

type preheatItem struct {
	authorID string
	postID   string
}

type API struct {
	PostRegistry       *search.PostRegistry
	LayoutServiceHost  string
	ThreadViewCacheTTL time.Duration
	ThreadViewCache    *lru.ARCCache
	LayoutCacheTTL     time.Duration
	LayoutCache        *lru.ARCCache
}

func main() {
	ctx := context.Background()
	var logger *zap.Logger

	if os.Getenv("DEBUG") == "true" {
		logger, _ = zap.NewDevelopment()
		logger.Info("Starting logger in DEBUG mode...")
	} else {
		logger, _ = zap.NewProduction()
		logger.Info("Starting logger in PRODUCTION mode...")
	}

	defer func() {
		err := logger.Sync()
		if err != nil {
			fmt.Printf("failed to sync logger on teardown: %+v", err.Error())
		}
	}()

	sugar := logger.Sugar()

	sugar.Info("Reading config from environment...")

	binReaderWriter := graph.BinaryGraphReaderWriter{}

	// Read the graph from the Binary file
	g1, err := binReaderWriter.ReadGraph("/app/graph.bin")
	if err != nil {
		log.Fatalf("Error reading graph1 from binary file: %v", err)
	}

	dbConnectionString := os.Getenv("REGISTRY_DB_CONNECTION_STRING")
	if dbConnectionString == "" {
		log.Fatal("REGISTRY_DB_CONNECTION_STRING environment variable is required")
	}

	layoutServiceHost := os.Getenv("LAYOUT_SERVICE_HOST")
	if layoutServiceHost == "" {
		layoutServiceHost = "http://localhost:8086"
	}

	// Registers a tracer Provider globally if the exporter endpoint is set
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		log.Println("initializing tracer...")
		shutdown, err := installExportPipeline(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Fatal(err)
			}
		}()
	}

	postRegistry, err := search.NewPostRegistry(dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to create PostRegistry: %v", err)
	}
	defer postRegistry.Close()

	// Hellthread is around 160KB right now so 1000 worst-case threads should be around 160MB
	threadViewCache, err := lru.NewARC(1000)
	if err != nil {
		log.Fatalf("Failed to create threadViewCache: %v", err)
	}

	layoutCache, err := lru.NewARC(500)
	if err != nil {
		log.Fatalf("Failed to create layoutCache: %v", err)
	}

	api := &API{
		PostRegistry:       postRegistry,
		LayoutServiceHost:  layoutServiceHost,
		ThreadViewCache:    threadViewCache,
		ThreadViewCacheTTL: 30 * time.Minute,
		LayoutCache:        layoutCache,
		LayoutCacheTTL:     30 * time.Minute,
	}

	router := gin.New()

	router.Use(gin.Recovery())

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			start := time.Now()
			// These can get consumed during request processing
			path := c.Request.URL.Path
			query := c.Request.URL.RawQuery
			c.Next()

			end := time.Now().UTC()
			latency := end.Sub(start)

			if len(c.Errors) > 0 {
				// Append error field if this is an erroneous request.
				for _, e := range c.Errors.Errors() {
					logger.Error(e)
				}
			} else if path != "/metrics" {
				logger.Info(path,
					zap.Int("status", c.Writer.Status()),
					zap.String("method", c.Request.Method),
					zap.String("path", path),
					zap.String("query", query),
					zap.String("ip", c.ClientIP()),
					zap.String("user-agent", c.Request.UserAgent()),
					zap.String("time", end.Format(time.RFC3339)),
					zap.String("rootPostID", c.GetString("rootPostID")),
					zap.String("rootPostAuthorDID", c.GetString("rootPostAuthorDID")),
					zap.Duration("latency", latency),
				)
			}
		}
	}())

	router.Use(ginzap.RecoveryWithZap(logger, true))

	router.Use(otelgin.Middleware("BSkySearchAPI"))

	// CORS middleware
	router.Use(cors.New(
		cors.Config{
			AllowOrigins: []string{"https://bsky.jazco.dev", "https://hellthread-explorer.bsky-graph.pages.dev"},
			AllowMethods: []string{"GET", "OPTIONS"},
			AllowHeaders: []string{"Origin", "Content-Length", "Content-Type"},
			AllowOriginFunc: func(origin string) bool {
				u, err := url.Parse(origin)
				if err != nil {
					return false
				}
				// Allow localhost and localnet requests for localdev
				return u.Hostname() == "localhost" || u.Hostname() == "10.0.6.32"
			},
		},
	))

	// Prometheus middleware
	p := ginprometheus.NewPrometheus("gin", nil)
	p.Use(router)

	router.GET("/thread", func(c *gin.Context) {
		authorID := c.Query("authorID")
		authorHandle := c.Query("authorHandle")
		postID := c.Query("postID")
		api.processThreadRequest(c, authorID, authorHandle, postID)
	})

	router.GET("/distance", func(c *gin.Context) {
		src := c.Query("src")
		dest := c.Query("dest")

		if src == "" || dest == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "src and dest must be provided"})
			return
		}

		// Make sure src and dst DIDs are in the graph
		if _, ok := g1.Nodes[graph.NodeID(src)]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("src with DID '%s' not found", src)})
			return
		}
		if _, ok := g1.Nodes[graph.NodeID(dest)]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("dest with DID '%s' not found", dest)})
			return
		}

		distance, path, weights := g1.FindSocialDistance(graph.NodeID(src), graph.NodeID(dest))

		// Return the distance, path, and weights with Handles and DIDs

		// Get the handles for the nodes in the path
		handles := make([]string, len(path))
		for i, nodeID := range path {
			handles[i] = g1.Nodes[nodeID].Handle
		}

		// Make sure weights and distnaces aren't infinite before returning them
		for i, weight := range weights {
			if math.IsInf(weight, 0) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("infinite weight for edge %s -> %s", path[i-1], path[i])})
				return
			}
		}

		if math.IsInf(distance, 0) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("infinite distance between %s and %s", src, dest)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"distance": distance, "did_path": path, "handle_path": handles, "weights": weights})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Preheat the caches with some popular threads
	preheatList := []preheatItem{
		{authorID: "did:plc:wgaezxqi2spqm3mhrb5xvkzi", postID: "3juzlwllznd24"},
	}

	// Create a routine to preheat the caches every 25 minutes

	go func() {
		ctx := context.Background()
		tracer := otel.Tracer("search-api")
		for {
			ctx, span := tracer.Start(ctx, "preheatCaches")
			log.Printf("Preheating caches with %d threads", len(preheatList))
			for _, threadToHeat := range preheatList {
				threadView, err := api.getThreadView(ctx, threadToHeat.postID, threadToHeat.authorID)
				if err != nil {
					log.Printf("Error preheating thread view cache: %v", err)
				}
				_, err = api.layoutThread(ctx, threadToHeat.postID, threadView)
				if err != nil {
					log.Printf("Error preheating layout cache: %v", err)
				}
			}
			span.End()
			// Sleep until the caches expire
			time.Sleep(30*time.Minute + 5*time.Second)
		}
	}()

	log.Printf("Starting server on port %s", port)
	router.Run(fmt.Sprintf(":%s", port))
}

func (api *API) layoutThread(ctx context.Context, rootPostID string, threadView []search.PostView) ([]layout.ThreadViewLayout, error) {
	tracer := otel.Tracer("search-api")
	ctx, span := tracer.Start(ctx, "layoutThread")
	defer span.End()

	// Check for the layout in the ARC Cache
	entry, ok := api.LayoutCache.Get(rootPostID)
	if ok {
		cacheEntry := entry.(LayoutCacheEntry)
		if cacheEntry.Expiration.After(time.Now()) {
			cacheHits.WithLabelValues("layout").Inc()
			span.SetAttributes(attribute.Bool("caches.layouts.hit", true))
			return cacheEntry.Layout, nil
		}
		// If the layout is expired, remove it from the cache
		api.LayoutCache.Remove(rootPostID)
	}

	cacheMisses.WithLabelValues("layout").Inc()

	threadViewLayout, err := layout.SendEdgeListRequestTS(ctx, api.LayoutServiceHost, threadView)
	if err != nil {
		return nil, fmt.Errorf("error sending edge list request: %w", err)
	}

	if threadViewLayout == nil {
		return nil, errors.New("layout service returned nil")
	}

	// Update the ARC Cache
	api.LayoutCache.Add(rootPostID, LayoutCacheEntry{
		Layout:     threadViewLayout,
		Expiration: time.Now().Add(api.LayoutCacheTTL),
	})

	return threadViewLayout, nil
}

func (api *API) processThreadRequest(c *gin.Context, authorID, authorHandle, postID string) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("search-api")
	ctx, span := tracer.Start(ctx, "processThreadRequest")
	defer span.End()
	span.SetAttributes(
		attribute.String("author.id", authorID),
		attribute.String("author.handle", authorHandle),
		attribute.String("post.id", postID),
	)

	if authorID == "" && authorHandle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorID or authorHandle must be provided"})
		return
	}

	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postID must be provided"})
		return
	}

	if authorID == "" {
		authors, err := api.PostRegistry.GetAuthorsByHandle(ctx, authorHandle)
		if err != nil {
			log.Printf("Error getting authors: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(authors) == 0 {
			log.Printf("Author with handle '%s' not found", authorHandle)
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Author with handle '%s' not found", authorHandle)})
			return
		}
		authorID = authors[0].DID
		span.SetAttributes(attribute.String("author.resolved_id", authorID))
	}

	// Get highest level post in thread
	rootPost, err := api.getRootOrOldestParent(ctx, postID)
	if err != nil {
		if errors.As(err, &search.NotFoundError{}) {
			log.Printf("Post with postID '%s' not found", postID)
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Post with postID '%s' not found", postID)})
		} else {
			log.Printf("Error getting root post: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if rootPost == nil {
		log.Printf("Post with postID '%s' not found", postID)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Post with postID '%s' not found", postID)})
		return
	}

	// Set the rootPostID in the context for the RequestLogger middleware
	c.Set("rootPostID", rootPost.ID)
	c.Set("rootPostAuthorDID", rootPost.AuthorDID)

	// Get thread view
	threadView, err := api.getThreadView(ctx, rootPost.ID, rootPost.AuthorDID)
	if err != nil {
		if errors.As(err, &search.NotFoundError{}) {
			log.Printf("Thread with authorID '%s' and postID '%s' not found", authorID, postID)
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Thread with authorID '%s' and postID '%s' not found", authorID, postID)})
		} else {
			log.Printf("Error getting thread view: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if c.Query("layout") == "true" {
		threadViewLayout, err := api.layoutThread(ctx, rootPost.ID, threadView)
		if err != nil {
			log.Printf("Error laying out thread: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, threadViewLayout)
		return
	}

	c.JSON(http.StatusOK, threadView)
}

func (api *API) getThreadView(ctx context.Context, postID, authorID string) ([]search.PostView, error) {
	tracer := otel.Tracer("search-api")
	ctx, span := tracer.Start(ctx, "getThreadView")
	defer span.End()

	// Check for the thread in the ARC Cache
	entry, ok := api.ThreadViewCache.Get(postID)
	if ok {
		cacheEntry := entry.(ThreadViewCacheEntry)
		if cacheEntry.Expiration.After(time.Now()) {
			cacheHits.WithLabelValues("thread").Inc()
			span.SetAttributes(attribute.Bool("caches.threads.hit", true))
			return cacheEntry.ThreadView, nil
		}
		// If the thread is expired, remove it from the cache
		api.ThreadViewCache.Remove(postID)
	}

	cacheMisses.WithLabelValues("thread").Inc()

	threadView, err := api.PostRegistry.GetThreadView(ctx, postID, authorID)
	if err != nil {
		if errors.As(err, &search.NotFoundError{}) {
			return nil, fmt.Errorf("thread with authorID '%s' and postID '%s' not found: %w", authorID, postID, err)
		}
		return nil, err
	}

	// Update the ARC Cache
	api.ThreadViewCache.Add(postID, ThreadViewCacheEntry{
		ThreadView: threadView,
		Expiration: time.Now().Add(api.ThreadViewCacheTTL),
	})

	return threadView, nil
}

func (api *API) getRootOrOldestParent(ctx context.Context, postID string) (*search.Post, error) {
	tracer := otel.Tracer("search-api")
	ctx, span := tracer.Start(ctx, "getRootOrOldestParent")
	defer span.End()
	// Get post from registry to look for root post
	span.AddEvent("getRootOrOldestParent:ResolvePrimaryPost")
	post, err := api.PostRegistry.GetPost(ctx, postID)
	if err != nil {
		if errors.As(err, &search.NotFoundError{}) {
			span.SetAttributes(attribute.Bool("post.primary.found", false))
			return nil, fmt.Errorf("post with postID '%s' not found: %w", postID, err)
		}
		return nil, err
	}

	span.SetAttributes(attribute.Bool("post.primary.found", true))

	// If post has a root post and we've stored it, return it
	if post.RootPostID != nil {
		span.AddEvent("getRootOrOldestParent:ResolveRootPost")
		rootPost, err := api.PostRegistry.GetPost(ctx, *post.RootPostID)
		if err != nil {
			// If we don't have the root post, continue to just return the oldest parent
			if !errors.As(err, &search.NotFoundError{}) {
				return nil, err
			}
			span.SetAttributes(attribute.Bool("post.root.found", false))
		}

		if rootPost != nil {
			span.SetAttributes(attribute.Bool("post.root.found", true))
			return rootPost, nil
		}
	}

	// Otherwise, get the oldest parent from the registry
	span.AddEvent("getRootOrOldestParent:ResolveOldestParent")
	oldestParent, err := api.PostRegistry.GetOldestPresentParent(ctx, postID)
	if err != nil {
		if errors.As(err, &search.NotFoundError{}) {
			span.SetAttributes(attribute.Bool("post.oldest_parent.found", false))
			return post, nil
		}
		return nil, err
	}

	if oldestParent != nil {
		span.SetAttributes(attribute.Bool("post.oldest_parent.found", true))
		return oldestParent, nil
	}

	return post, nil
}

func installExportPipeline(ctx context.Context) (func(context.Context) error, error) {
	client := otlptracehttp.NewClient()
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	tracerProvider := newTraceProvider(exporter)
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("BSkySearchAPI"),
		),
	)

	if err != nil {
		panic(err)
	}

	// initialize the traceIDRatioBasedSampler
	traceIDRatioBasedSampler := sdktrace.TraceIDRatioBased(1)

	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(traceIDRatioBasedSampler),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

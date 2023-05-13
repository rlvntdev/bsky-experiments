package usercount

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	intXRPC "github.com/ericvolp12/bsky-experiments/pkg/xrpc"
	"go.opentelemetry.io/otel"

	"golang.org/x/time/rate"
)

type UserCount struct {
	Client           *xrpc.Client
	ClientMux        *sync.RWMutex
	CurrentUserCount int
	RateLimiter      *rate.Limiter
	Cursors          []string
}

func NewUserCount(ctx context.Context, client *xrpc.Client) *UserCount {
	clientMux := &sync.RWMutex{}

	// Run a routine that refreshes the auth token every 10 minutes
	authTicker := time.NewTicker(10 * time.Minute)
	quit := make(chan struct{})
	go func() {
		log.Println("starting auth refresh routine...")
		for {
			select {
			case <-authTicker.C:
				log.Println("refreshing auth token...")
				err := intXRPC.RefreshAuth(ctx, client, clientMux)
				if err != nil {
					log.Printf("error refreshing auth token: %s\n", err)
				} else {
					log.Println("successfully refreshed auth token")
				}
			case <-quit:
				authTicker.Stop()
				return
			}
		}
	}()

	// Set up a rate limiter to limit requests to 5 per second
	limiter := rate.NewLimiter(rate.Limit(5), 1)

	return &UserCount{
		Client:      client,
		RateLimiter: limiter,
		ClientMux:   clientMux,
		Cursors:     []string{""},
	}
}

// GetUserCount returns the number of users of BSky from the Repo Sync API
// It uses a rate limiter to limit requests to 5 per second
// It does not implement any caching, so it will make a series of requests to the API every time it is called
// Caching should be implemented one layer up in the application
func (uc *UserCount) GetUserCount(ctx context.Context) (int, error) {
	tracer := otel.Tracer("usercount")
	ctx, span := tracer.Start(ctx, "GetUserCount")
	defer span.End()

	numUsers := 0

	countChannel := make(chan int, len(uc.Cursors)-1)
	errorChannel := make(chan error, len(uc.Cursors)-1)

	// Iterate over all but the final cursor in goroutines
	for i, cursor := range uc.Cursors {
		if i == len(uc.Cursors)-1 {
			break
		}
		go func(cursor string) {
			log.Printf("getting repos with cursor %s\n", cursor)

			// Use rate limiter before each request
			err := uc.RateLimiter.Wait(ctx)
			if err != nil {
				fmt.Printf("error waiting for rate limiter: %v", err)
				errorChannel <- fmt.Errorf("error waiting for rate limiter: %w", err)
				return
			}

			span.AddEvent("AcquireClientRLock")
			uc.ClientMux.RLock()
			span.AddEvent("ClientRLockAcquired")
			repoOutput, err := comatproto.SyncListRepos(ctx, uc.Client, cursor, 1000)
			if err != nil {
				fmt.Printf("error listing repos: %s\n", err)
				errorChannel <- fmt.Errorf("error listing repos: %w", err)
				return
			}
			span.AddEvent("ReleaseClientRLock")
			uc.ClientMux.RUnlock()

			countChannel <- len(repoOutput.Repos)
		}(cursor)
	}

	// Read from the channels
	for i := 0; i < len(uc.Cursors)-1; i++ {
		select {
		case count := <-countChannel:
			numUsers += count
		case err := <-errorChannel:
			return -1, err
		}
	}

	log.Printf("fetching final cursor: %s\n", uc.Cursors[len(uc.Cursors)-1])

	cursor := uc.Cursors[len(uc.Cursors)-1]
	// Iterate over the final cursor in the main thread
	// this also builds up the cursor list on the first run
	for {
		// Use rate limiter before each request
		err := uc.RateLimiter.Wait(ctx)
		if err != nil {
			fmt.Printf("error waiting for rate limiter: %v", err)
			return -1, fmt.Errorf("error waiting for rate limiter: %w", err)
		}

		span.AddEvent("AcquireClientRLock")
		uc.ClientMux.RLock()
		span.AddEvent("ClientRLockAcquired")
		repoOutput, err := comatproto.SyncListRepos(ctx, uc.Client, cursor, 1000)
		if err != nil {
			fmt.Printf("error listing repos: %s\n", err)
			return -1, fmt.Errorf("error listing repos: %w", err)
		}
		span.AddEvent("ReleaseClientRLock")
		uc.ClientMux.RUnlock()

		numUsers += len(repoOutput.Repos)

		if repoOutput.Cursor == nil {
			break
		}

		cursor = *repoOutput.Cursor
		uc.Cursors = append(uc.Cursors, cursor)
	}

	uc.CurrentUserCount = numUsers

	return numUsers, nil
}

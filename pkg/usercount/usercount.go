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

	"golang.org/x/time/rate"
)

type UserCount struct {
	Client           *xrpc.Client
	ClientMux        *sync.RWMutex
	CurrentUserCount int
	RateLimiter      *rate.Limiter
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
	}
}

// GetUserCount returns the number of users of BSky from the Repo Sync API
// It uses a rate limiter to limit requests to 5 per second
// It does not implement any caching, so it will make a series of requests to the API every time it is called
// Caching should be implemented one layer up in the application
func (uc *UserCount) GetUserCount(ctx context.Context) (int, error) {
	cursor := ""

	numUsers := 0

	for {
		// Use rate limiter before each request
		err := uc.RateLimiter.Wait(ctx)
		if err != nil {
			fmt.Printf("error waiting for rate limiter: %v", err)
			return -1, fmt.Errorf("error waiting for rate limiter: %w", err)
		}

		repoOutput, err := comatproto.SyncListRepos(ctx, uc.Client, cursor, 1000)
		if err != nil {
			fmt.Printf("error listing repos: %s\n", err)
			return -1, fmt.Errorf("error listing repos: %w", err)
		}

		numUsers += len(repoOutput.Repos)

		if repoOutput.Cursor == nil {
			break
		}

		cursor = *repoOutput.Cursor
	}

	uc.CurrentUserCount = numUsers

	return numUsers, nil
}
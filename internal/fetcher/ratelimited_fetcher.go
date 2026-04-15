package fetcher

import (
	"context"
	"fmt"

	"github.com/example/driftwatch/internal/ratelimit"
)

// Fetcher is the interface satisfied by HTTP and cached fetchers.
type Fetcher interface {
	Fetch(ctx context.Context, url string) (map[string]interface{}, error)
}

// RateLimitedFetcher wraps a Fetcher and enforces a rate limit on calls.
type RateLimitedFetcher struct {
	inner   Fetcher
	limiter *ratelimit.Limiter
}

// NewRateLimited wraps inner with the provided Limiter.
// Returns an error if inner or limiter is nil.
func NewRateLimited(inner Fetcher, limiter *ratelimit.Limiter) (*RateLimitedFetcher, error) {
	if inner == nil {
		return nil, fmt.Errorf("ratelimited_fetcher: inner fetcher must not be nil")
	}
	if limiter == nil {
		return nil, fmt.Errorf("ratelimited_fetcher: limiter must not be nil")
	}
	return &RateLimitedFetcher{inner: inner, limiter: limiter}, nil
}

// Fetch waits for a rate-limit token then delegates to the inner Fetcher.
func (r *RateLimitedFetcher) Fetch(ctx context.Context, url string) (map[string]interface{}, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("ratelimited_fetcher: rate limit wait: %w", err)
	}
	return r.inner.Fetch(ctx, url)
}

package fetcher

import (
	"context"
	"fmt"
	"time"
)

// Fetcher is the interface expected by RetryFetcher.
type retryable interface {
	Fetch(ctx context.Context, url string) (map[string]interface{}, error)
}

// RetryFetcher wraps another Fetcher and retries on transient errors.
type RetryFetcher struct {
	inner    retryable
	maxRetry int
	delay    time.Duration
}

// NewRetry creates a RetryFetcher.
// maxRetry is the number of additional attempts after the first failure.
// delay is the wait time between attempts.
func NewRetry(inner retryable, maxRetry int, delay time.Duration) (*RetryFetcher, error) {
	if inner == nil {
		return nil, fmt.Errorf("retry fetcher: inner fetcher must not be nil")
	}
	if maxRetry < 1 {
		return nil, fmt.Errorf("retry fetcher: maxRetry must be at least 1, got %d", maxRetry)
	}
	if delay < 0 {
		return nil, fmt.Errorf("retry fetcher: delay must be non-negative")
	}
	return &RetryFetcher{inner: inner, maxRetry: maxRetry, delay: delay}, nil
}

// Fetch attempts to fetch from the inner fetcher, retrying up to maxRetry times.
func (r *RetryFetcher) Fetch(ctx context.Context, url string) (map[string]interface{}, error) {
	var lastErr error
	for attempt := 0; attempt <= r.maxRetry; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("retry fetcher: context cancelled after %d attempt(s): %w", attempt, ctx.Err())
			case <-time.After(r.delay):
			}
		}
		result, err := r.inner.Fetch(ctx, url)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("retry fetcher: all %d attempt(s) failed for %s: %w", r.maxRetry+1, url, lastErr)
}

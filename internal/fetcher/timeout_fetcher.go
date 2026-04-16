package fetcher

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TimeoutFetcher wraps a Fetcher and enforces a per-request deadline.
type TimeoutFetcher struct {
	inner   Fetcher
	timeout time.Duration
}

// NewTimeout creates a TimeoutFetcher. timeout must be positive.
func NewTimeout(inner Fetcher, timeout time.Duration) (*TimeoutFetcher, error) {
	if inner == nil {
		return nil, errors.New("timeout fetcher: inner fetcher must not be nil")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("timeout fetcher: timeout must be positive, got %s", timeout)
	}
	return &TimeoutFetcher{inner: inner, timeout: timeout}, nil
}

// Fetch calls the inner fetcher with a bounded context.
func (t *TimeoutFetcher) Fetch(ctx context.Context, url string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	result, err := t.inner.Fetch(ctx, url)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("timeout fetcher: request to %s exceeded %s", url, t.timeout)
		}
		return nil, err
	}
	return result, nil
}

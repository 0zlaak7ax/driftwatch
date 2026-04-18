package fetcher

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// BatchFetcher fetches multiple URLs concurrently and returns a map of results.
type BatchFetcher struct {
	inner   Fetcher
	workers int
}

// BatchResult holds the result for a single URL.
type BatchResult struct {
	URL  string
	Data map[string]any
	Err  error
}

// NewBatch creates a BatchFetcher wrapping the given Fetcher.
func NewBatch(inner Fetcher, workers int) (*BatchFetcher, error) {
	if inner == nil {
		return nil, errors.New("batch: inner fetcher must not be nil")
	}
	if workers < 1 {
		return nil, fmt.Errorf("batch: workers must be >= 1, got %d", workers)
	}
	return &BatchFetcher{inner: inner, workers: workers}, nil
}

// FetchAll fetches all given URLs concurrently and returns a slice of BatchResult.
func (b *BatchFetcher) FetchAll(ctx context.Context, urls []string) []BatchResult {
	results := make([]BatchResult, len(urls))
	sem := make(chan struct{}, b.workers)
	var wg sync.WaitGroup

	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			data, err := b.inner.Fetch(ctx, u)
			results[idx] = BatchResult{URL: u, Data: data, Err: err}
		}(i, url)
	}

	wg.Wait()
	return results
}

package fetcher

import (
	"context"
	"errors"
	"sync"
)

// Fetcher is the interface for fetching live service config.
type poolFetcher struct {
	workers int
	inner   Fetcher
}

// FetchResult holds the result of a single fetch.
type FetchResult struct {
	URL    string
	Data   map[string]interface{}
	Err    error
}

// NewPool returns a fetcher that concurrently fetches multiple URLs using inner,
// bounded by the given worker count.
func NewPool(inner Fetcher, workers int) (*poolFetcher, error) {
	if inner == nil {
		return nil, errors.New("pool: inner fetcher must not be nil")
	}
	if workers < 1 {
		return nil, errors.New("pool: workers must be at least 1")
	}
	return &poolFetcher{inner: inner, workers: workers}, nil
}

// FetchAll fetches all provided URLs concurrently and returns results in input order.
func (p *poolFetcher) FetchAll(ctx context.Context, urls []string) []FetchResult {
	results := make([]FetchResult, len(urls))
	sem := make(chan struct{}, p.workers)
	var wg sync.WaitGroup

	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			data, err := p.inner.Fetch(ctx, u)
			results[idx] = FetchResult{URL: u, Data: data, Err: err}
		}(i, url)
	}

	wg.Wait()
	return results
}

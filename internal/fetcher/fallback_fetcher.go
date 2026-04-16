package fetcher

import (
	"context"
	"fmt"
)

// FallbackFetcher tries a primary fetcher and falls back to a secondary on error.
type FallbackFetcher struct {
	primary  Fetcher
	fallback Fetcher
}

// NewFallback creates a FallbackFetcher. Both primary and fallback must be non-nil.
func NewFallback(primary, fallback Fetcher) (*FallbackFetcher, error) {
	if primary == nil {
		return nil, fmt.Errorf("fallback fetcher: primary must not be nil")
	}
	if fallback == nil {
		return nil, fmt.Errorf("fallback fetcher: fallback must not be nil")
	}
	return &FallbackFetcher{primary: primary, fallback: fallback}, nil
}

// Fetch attempts the primary fetcher; on any error it delegates to the fallback.
func (f *FallbackFetcher) Fetch(ctx context.Context, url string) (map[string]interface{}, error) {
	result, err := f.primary.Fetch(ctx, url)
	if err == nil {
		return result, nil
	}
	return f.fallback.Fetch(ctx, url)
}

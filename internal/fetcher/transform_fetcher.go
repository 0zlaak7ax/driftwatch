package fetcher

import (
	"errors"
	"fmt"
)

// TransformFunc mutates a live config map after fetching.
type TransformFunc func(map[string]string) (map[string]string, error)

// transformFetcher wraps an inner Fetcher and applies a transformation to results.
type transformFetcher struct {
	inner     Fetcher
	transform TransformFunc
}

// NewTransform returns a Fetcher that applies fn to every fetched config.
func NewTransform(inner Fetcher, fn TransformFunc) (Fetcher, error) {
	if inner == nil {
		return nil, errors.New("transform fetcher: inner fetcher must not be nil")
	}
	if fn == nil {
		return nil, errors.New("transform fetcher: transform func must not be nil")
	}
	return &transformFetcher{inner: inner, transform: fn}, nil
}

func (t *transformFetcher) Fetch(url string) (map[string]string, error) {
	data, err := t.inner.Fetch(url)
	if err != nil {
		return nil, err
	}
	result, err := t.transform(data)
	if err != nil {
		return nil, fmt.Errorf("transform fetcher: transform failed: %w", err)
	}
	return result, nil
}

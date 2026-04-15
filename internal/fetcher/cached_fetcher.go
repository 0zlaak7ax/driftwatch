package fetcher

import (
	"fmt"
	"time"

	"github.com/yourorg/driftwatch/internal/cache"
)

// Fetcher is the interface for retrieving live service state.
type Fetcher interface {
	Fetch(url string) (map[string]interface{}, error)
}

// CachedFetcher wraps a Fetcher with an in-memory TTL cache.
type CachedFetcher struct {
	inner Fetcher
	cache *cache.Cache
}

// NewCached returns a CachedFetcher that caches results from inner for ttl.
func NewCached(inner Fetcher, ttl time.Duration) *CachedFetcher {
	return &CachedFetcher{
		inner: inner,
		cache: cache.New(ttl),
	}
}

// Fetch returns a cached result if available and unexpired; otherwise it
// delegates to the inner Fetcher, stores the result, and returns it.
func (cf *CachedFetcher) Fetch(url string) (map[string]interface{}, error) {
	if data, ok := cf.cache.Get(url); ok {
		return data, nil
	}

	data, err := cf.inner.Fetch(url)
	if err != nil {
		return nil, fmt.Errorf("cached fetcher: %w", err)
	}

	cf.cache.Set(url, data)
	return data, nil
}

// Invalidate removes the cached entry for url, forcing a fresh fetch next time.
func (cf *CachedFetcher) Invalidate(url string) {
	cf.cache.Invalidate(url)
}

package fetcher

import (
	"errors"
	"sync"
)

// DedupeFetcher coalesces concurrent in-flight requests for the same URL
// into a single upstream fetch.
type DedupeFetcher struct {
	inner  Fetcher
	mu     sync.Mutex
	flying map[string]*call
}

type call struct {
	wg  sync.WaitGroup
	res map[string]interface{}
	err error
}

// NewDedupe wraps inner with request deduplication.
func NewDedupe(inner Fetcher) (*DedupeFetcher, error) {
	if inner == nil {
		return nil, errors.New("dedupe: inner fetcher must not be nil")
	}
	return &DedupeFetcher{
		inner:  inner,
		flying: make(map[string]*call),
	}, nil
}

// Fetch delegates to inner, deduplicating concurrent calls for the same url.
// If multiple goroutines call Fetch with the same url simultaneously, only one
// upstream request is made; the rest wait and share the result.
func (d *DedupeFetcher) Fetch(url string) (map[string]interface{}, error) {
	d.mu.Lock()
	if c, ok := d.flying[url]; ok {
		d.mu.Unlock()
		c.wg.Wait()
		return c.res, c.err
	}
	c := &call{}
	c.wg.Add(1)
	d.flying[url] = c
	d.mu.Unlock()

	c.res, c.err = d.inner.Fetch(url)
	c.wg.Done()

	d.mu.Lock()
	delete(d.flying, url)
	d.mu.Unlock()

	return c.res, c.err
}

// InFlight returns the number of URLs currently being fetched.
func (d *DedupeFetcher) InFlight() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.flying)
}

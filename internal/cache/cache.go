package cache

import (
	"sync"
	"time"
)

// Entry holds a cached fetch result with an expiry timestamp.
type Entry struct {
	Data      map[string]interface{}
	FetchedAt time.Time
	ExpiresAt time.Time
}

// Cache is a simple in-memory TTL cache for fetched service state.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]Entry
	ttl     time.Duration
}

// New creates a Cache with the given TTL duration.
func New(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]Entry),
		ttl:     ttl,
	}
}

// Set stores data for the given key, overwriting any existing entry.
func (c *Cache) Set(key string, data map[string]interface{}) {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = Entry{
		Data:      data,
		FetchedAt: now,
		ExpiresAt: now.Add(c.ttl),
	}
}

// Get retrieves data for the given key. Returns (data, true) if the entry
// exists and has not expired, otherwise (nil, false).
func (c *Cache) Get(key string) (map[string]interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Data, true
}

// Invalidate removes the entry for the given key.
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Purge removes all expired entries from the cache.
func (c *Cache) Purge() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, e := range c.entries {
		if now.After(e.ExpiresAt) {
			delete(c.entries, k)
		}
	}
}

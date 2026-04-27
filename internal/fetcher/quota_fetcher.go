package fetcher

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// QuotaFetcher wraps a Fetcher and enforces a maximum number of fetches
// per time window (quota). Unlike the rate limiter, quota resets on a
// fixed wall-clock boundary rather than a sliding window.
type QuotaFetcher struct {
	inner    Fetcher
	max      int
	window   time
	mu       sync.Mutex
	count    int
	windowAt time.Time
}

var ErrQuotaExceeded = errors.New("quota_fetcher: quota exceeded for this window")

// NewQuota returns a QuotaFetcher that allows at most max fetches per window
// duration. Returns an error if inner is nil, max < 1, or window <= 0.
func NewQuota(inner Fetcher, max int, window time.Duration) (*QuotaFetcher, error) {
	if inner == nil {
		return nil, errors.New("quota_fetcher: inner fetcher must not be nil")
	}
	if max < 1 {
		return nil, fmt.Errorf("quota_fetcher: max must be >= 1, got %d", max)
	}
	if window <= 0 {
		return nil, fmt.Errorf("quota_fetcher: window must be positive, got %s", window)
	}
	return &QuotaFetcher{
		inner:    inner,
		max:      max,
		window:   window,
		windowAt: time.Now(),
	}, nil
}

// Fetch delegates to the inner fetcher if the quota has not been exhausted
// for the current window. The window resets automatically when it expires.
func (q *QuotaFetcher) Fetch(service, url string) (map[string]interface{}, error) {
	q.mu.Lock()
	now := time.Now()
	if now.After(q.windowAt.Add(q.window)) {
		q.count = 0
		q.windowAt = now
	}
	if q.count >= q.max {
		q.mu.Unlock()
		return nil, fmt.Errorf("%w: service=%s limit=%d window=%s",
			ErrQuotaExceeded, service, q.max, q.window)
	}
	q.count++
	q.mu.Unlock()
	return q.inner.Fetch(service, url)
}

// Remaining returns the number of fetches still allowed in the current window.
func (q *QuotaFetcher) Remaining() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	if time.Now().After(q.windowAt.Add(q.window)) {
		return q.max
	}
	rem := q.max - q.count
	if rem < 0 {
		return 0
	}
	return rem
}

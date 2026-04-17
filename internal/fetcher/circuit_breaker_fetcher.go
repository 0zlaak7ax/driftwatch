package fetcher

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreakerFetcher wraps a Fetcher with a circuit breaker.
type CircuitBreakerFetcher struct {
	inner      Fetcher
	threshold  int
	resetAfter time.Duration

	mu        sync.Mutex
	failures  int
	state     state
	openedAt  time.Time
}

// NewCircuitBreaker creates a CircuitBreakerFetcher.
// threshold is the number of consecutive failures before opening.
// resetAfter is how long to wait before attempting half-open.
func NewCircuitBreaker(inner Fetcher, threshold int, resetAfter time.Duration) (*CircuitBreakerFetcher, error) {
	if inner == nil {
		return nil, errors.New("inner fetcher must not be nil")
	}
	if threshold < 1 {
		return nil, fmt.Errorf("threshold must be >= 1, got %d", threshold)
	}
	if resetAfter <= 0 {
		return nil, fmt.Errorf("resetAfter must be positive")
	}
	return &CircuitBreakerFetcher{
		inner:      inner,
		threshold:  threshold,
		resetAfter: resetAfter,
		state:      stateClosed,
	}, nil
}

// Fetch delegates to the inner fetcher, applying circuit breaker logic.
func (cb *CircuitBreakerFetcher) Fetch(url string) (map[string]interface{}, error) {
	cb.mu.Lock()
	switch cb.state {
	case stateOpen:
		if time.Since(cb.openedAt) >= cb.resetAfter {
			cb.state = stateHalfOpen
		} else {
			cb.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	case stateHalfOpen, stateClosed:
		// allow through
	}
	cb.mu.Unlock()

	result, err := cb.inner.Fetch(url)

	cb.mu.Lock()
	defer cb.mu.Unlock()
	if err != nil {
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.state = stateOpen
			cb.openedAt = time.Now()
		}
		return nil, err
	}
	cb.failures = 0
	cb.state = stateClosed
	return result, nil
}

package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Limiter enforces a maximum number of operations per interval.
type Limiter struct {
	mu       sync.Mutex
	rate     int
	interval time.Duration
	tokens   int
	lastFill time.Time
}

// Config holds the configuration for a Limiter.
type Config struct {
	// Rate is the maximum number of allowed operations per Interval.
	Rate     int
	Interval time.Duration
}

// New creates a new Limiter from the given Config.
// Returns an error if Rate <= 0 or Interval <= 0.
func New(cfg Config) (*Limiter, error) {
	if cfg.Rate <= 0 {
		return nil, fmt.Errorf("ratelimit: rate must be > 0, got %d", cfg.Rate)
	}
	if cfg.Interval <= 0 {
		return nil, fmt.Errorf("ratelimit: interval must be > 0, got %s", cfg.Interval)
	}
	return &Limiter{
		rate:     cfg.Rate,
		interval: cfg.Interval,
		tokens:   cfg.Rate,
		lastFill: time.Now(),
	}, nil
}

// Allow reports whether an operation is permitted right now.
// It refills tokens based on elapsed time before checking.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if l.tokens > 0 {
		l.tokens--
		return true
	}
	return false
}

// Wait blocks until an operation is permitted or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if l.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.interval / time.Duration(l.rate)):
		}
	}
}

// refill adds tokens based on elapsed time since the last fill.
// Must be called with l.mu held.
func (l *Limiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastFill)
	if elapsed >= l.interval {
		periods := int(elapsed / l.interval)
		l.tokens += periods * l.rate
		if l.tokens > l.rate {
			l.tokens = l.rate
		}
		l.lastFill = l.lastFill.Add(time.Duration(periods) * l.interval)
	}
}

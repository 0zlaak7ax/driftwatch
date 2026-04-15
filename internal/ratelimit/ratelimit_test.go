package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/ratelimit"
)

func TestNew_Valid(t *testing.T) {
	_, err := ratelimit.New(ratelimit.Config{Rate: 5, Interval: time.Second})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNew_InvalidRate(t *testing.T) {
	_, err := ratelimit.New(ratelimit.Config{Rate: 0, Interval: time.Second})
	if err == nil {
		t.Fatal("expected error for rate=0")
	}
}

func TestNew_InvalidInterval(t *testing.T) {
	_, err := ratelimit.New(ratelimit.Config{Rate: 1, Interval: 0})
	if err == nil {
		t.Fatal("expected error for interval=0")
	}
}

func TestAllow_WithinRate(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.Config{Rate: 3, Interval: time.Second})
	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsRate(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.Config{Rate: 2, Interval: time.Second})
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("expected Allow()=false after exhausting tokens")
	}
}

func TestAllow_RefillsAfterInterval(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.Config{Rate: 1, Interval: 50 * time.Millisecond})
	if !l.Allow() {
		t.Fatal("expected first Allow()=true")
	}
	if l.Allow() {
		t.Fatal("expected second Allow()=false before refill")
	}
	time.Sleep(60 * time.Millisecond)
	if !l.Allow() {
		t.Fatal("expected Allow()=true after refill interval")
	}
}

func TestWait_PermitsEventually(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.Config{Rate: 1, Interval: 50 * time.Millisecond})
	l.Allow() // exhaust
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("expected Wait to succeed, got %v", err)
	}
}

func TestWait_CancelledContext(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.Config{Rate: 1, Interval: 10 * time.Second})
	l.Allow() // exhaust
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.Wait(ctx); err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

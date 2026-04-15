package fetcher_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/driftwatch/internal/fetcher"
)

type countingFetcher struct {
	calls     int
	failUntil int
	result    map[string]interface{}
}

func (c *countingFetcher) Fetch(_ context.Context, _ string) (map[string]interface{}, error) {
	c.calls++
	if c.calls <= c.failUntil {
		return nil, fmt.Errorf("transient error attempt %d", c.calls)
	}
	return c.result, nil
}

func TestNewRetry_NilInner(t *testing.T) {
	_, err := fetcher.NewRetry(nil, 3, 0)
	if err == nil {
		t.Fatal("expected error for nil inner fetcher")
	}
}

func TestNewRetry_InvalidMaxRetry(t *testing.T) {
	f := &countingFetcher{}
	_, err := fetcher.NewRetry(f, 0, 0)
	if err == nil {
		t.Fatal("expected error for maxRetry < 1")
	}
}

func TestNewRetry_NegativeDelay(t *testing.T) {
	f := &countingFetcher{}
	_, err := fetcher.NewRetry(f, 2, -time.Millisecond)
	if err == nil {
		t.Fatal("expected error for negative delay")
	}
}

func TestRetry_SucceedsFirstAttempt(t *testing.T) {
	inner := &countingFetcher{failUntil: 0, result: map[string]interface{}{"status": "ok"}}
	rf, _ := fetcher.NewRetry(inner, 3, 0)
	res, err := rf.Fetch(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["status"] != "ok" {
		t.Errorf("unexpected result: %v", res)
	}
	if inner.calls != 1 {
		t.Errorf("expected 1 call, got %d", inner.calls)
	}
}

func TestRetry_SucceedsAfterRetries(t *testing.T) {
	inner := &countingFetcher{failUntil: 2, result: map[string]interface{}{"env": "prod"}}
	rf, _ := fetcher.NewRetry(inner, 3, 0)
	res, err := rf.Fetch(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["env"] != "prod" {
		t.Errorf("unexpected result: %v", res)
	}
	if inner.calls != 3 {
		t.Errorf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetry_ExhaustsAllAttempts(t *testing.T) {
	inner := &countingFetcher{failUntil: 99}
	rf, _ := fetcher.NewRetry(inner, 2, 0)
	_, err := rf.Fetch(context.Background(), "http://example.com")
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if inner.calls != 3 {
		t.Errorf("expected 3 total calls (1 + 2 retries), got %d", inner.calls)
	}
}

func TestRetry_ContextCancelled(t *testing.T) {
	inner := &countingFetcher{failUntil: 99}
	rf, _ := fetcher.NewRetry(inner, 5, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := rf.Fetch(ctx, "http://example.com")
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled in error chain, got: %v", err)
	}
}

package fetcher_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/fetcher"
	"github.com/example/driftwatch/internal/ratelimit"
)

type stubFetcher struct {
	result map[string]interface{}
	err    error
	calls  int
}

func (s *stubFetcher) Fetch(_ context.Context, _ string) (map[string]interface{}, error) {
	s.calls++
	return s.result, s.err
}

func newLimiter(t *testing.T, rate int, interval time.Duration) *ratelimit.Limiter {
	t.Helper()
	l, err := ratelimit.New(ratelimit.Config{Rate: rate, Interval: interval})
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}
	return l
}

func TestNewRateLimited_NilInner(t *testing.T) {
	l := newLimiter(t, 5, time.Second)
	_, err := fetcher.NewRateLimited(nil, l)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewRateLimited_NilLimiter(t *testing.T) {
	_, err := fetcher.NewRateLimited(&stubFetcher{}, nil)
	if err == nil {
		t.Fatal("expected error for nil limiter")
	}
}

func TestRateLimited_Fetch_Success(t *testing.T) {
	expected := map[string]interface{}{"version": "1.2.3"}
	stub := &stubFetcher{result: expected}
	l := newLimiter(t, 5, time.Second)
	rf, err := fetcher.NewRateLimited(stub, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := rf.Fetch(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got["version"] != "1.2.3" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestRateLimited_Fetch_PropagatesInnerError(t *testing.T) {
	stub := &stubFetcher{err: errors.New("connection refused")}
	l := newLimiter(t, 5, time.Second)
	rf, _ := fetcher.NewRateLimited(stub, l)
	_, err := rf.Fetch(context.Background(), "http://example.com")
	if err == nil {
		t.Fatal("expected error from inner fetcher")
	}
}

func TestRateLimited_Fetch_CancelledContext(t *testing.T) {
	stub := &stubFetcher{}
	l := newLimiter(t, 1, 10*time.Second)
	l.Allow() // exhaust tokens
	rf, _ := fetcher.NewRateLimited(stub, l)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := rf.Fetch(ctx, "http://example.com")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if stub.calls != 0 {
		t.Errorf("inner fetcher should not have been called, got %d calls", stub.calls)
	}
}

package fetcher_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/driftwatch/internal/fetcher"
)

type slowFetcher struct {
	delay time.Duration
}

func (s *slowFetcher) Fetch(ctx context.Context, url string) (map[string]interface{}, error) {
	select {
	case <-time.After(s.delay):
		return map[string]interface{}{"status": "ok"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestNewTimeout_NilInner(t *testing.T) {
	_, err := fetcher.NewTimeout(nil, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner fetcher")
	}
}

func TestNewTimeout_InvalidDuration(t *testing.T) {
	sf := &slowFetcher{}
	_, err := fetcher.NewTimeout(sf, 0)
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
	_, err = fetcher.NewTimeout(sf, -time.Second)
	if err == nil {
		t.Fatal("expected error for negative timeout")
	}
}

func TestTimeout_Fetch_Success(t *testing.T) {
	sf := &slowFetcher{delay: 10 * time.Millisecond}
	tf, err := fetcher.NewTimeout(sf, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := tf.Fetch(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestTimeout_Fetch_Exceeded(t *testing.T) {
	sf := &slowFetcher{delay: 300 * time.Millisecond}
	tf, err := fetcher.NewTimeout(sf, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = tf.Fetch(context.Background(), "http://example.com")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestTimeout_Fetch_InnerError(t *testing.T) {
	errFetcher := &stubFetcher{err: errors.New("inner error")}
	tf, err := fetcher.NewTimeout(errFetcher, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = tf.Fetch(context.Background(), "http://example.com")
	if err == nil {
		t.Fatal("expected inner error")
	}
}

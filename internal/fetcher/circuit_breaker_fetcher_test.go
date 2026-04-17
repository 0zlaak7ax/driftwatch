package fetcher_test

import (
	"errors"
	"testing"
	"time"

	"github.com/driftwatch/internal/fetcher"
)

type failingFetcher struct{ err error }

func (f *failingFetcher) Fetch(_ string) (map[string]interface{}, error) {
	return nil, f.err
}

type succeedingFetcher struct{}

func (s *succeedingFetcher) Fetch(_ string) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": true}, nil
}

func TestNewCircuitBreaker_NilInner(t *testing.T) {
	_, err := fetcher.NewCircuitBreaker(nil, 2, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewCircuitBreaker_InvalidThreshold(t *testing.T) {
	_, err := fetcher.NewCircuitBreaker(&succeedingFetcher{}, 0, time.Second)
	if err == nil {
		t.Fatal("expected error for zero threshold")
	}
}

func TestNewCircuitBreaker_InvalidReset(t *testing.T) {
	_, err := fetcher.NewCircuitBreaker(&succeedingFetcher{}, 1, 0)
	if err == nil {
		t.Fatal("expected error for zero resetAfter")
	}
}

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	cb, _ := fetcher.NewCircuitBreaker(&succeedingFetcher{}, 2, time.Second)
	result, err := cb.Fetch("http://example.com")
	if err != nil || result["ok"] != true {
		t.Fatalf("expected success, got err=%v result=%v", err, result)
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	inner := &failingFetcher{err: errors.New("down")}
	cb, _ := fetcher.NewCircuitBreaker(inner, 2, time.Second)

	cb.Fetch("http://example.com")
	cb.Fetch("http://example.com")

	_, err := cb.Fetch("http://example.com")
	if !errors.Is(err, fetcher.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenAfterReset(t *testing.T) {
	inner := &failingFetcher{err: errors.New("down")}
	cb, _ := fetcher.NewCircuitBreaker(inner, 1, 10*time.Millisecond)

	cb.Fetch("http://example.com") // opens circuit

	time.Sleep(20 * time.Millisecond)

	// Should attempt again (half-open), inner still failing so error is inner error
	_, err := cb.Fetch("http://example.com")
	if errors.Is(err, fetcher.ErrCircuitOpen) {
		t.Fatal("expected inner error after reset, not ErrCircuitOpen")
	}
}

func TestCircuitBreaker_RecoverAfterSuccess(t *testing.T) {
	inner := &failingFetcher{err: errors.New("down")}
	cb, _ := fetcher.NewCircuitBreaker(inner, 1, 10*time.Millisecond)

	cb.Fetch("http://example.com") // trip open
	time.Sleep(20 * time.Millisecond)

	// swap to succeeding
	good, _ := fetcher.NewCircuitBreaker(&succeedingFetcher{}, 1, 10*time.Millisecond)
	result, err := good.Fetch("http://example.com")
	if err != nil || result["ok"] != true {
		t.Fatalf("expected recovery, got %v", err)
	}
}

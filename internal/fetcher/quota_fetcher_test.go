package fetcher_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/fetcher"
)

func TestNewQuota_NilInner(t *testing.T) {
	_, err := fetcher.NewQuota(nil, 5, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewQuota_InvalidMax(t *testing.T) {
	f, _ := fetcher.New(fetcher.Options{})
	_, err := fetcher.NewQuota(f, 0, time.Second)
	if err == nil {
		t.Fatal("expected error for max=0")
	}
}

func TestNewQuota_InvalidWindow(t *testing.T) {
	f, _ := fetcher.New(fetcher.Options{})
	_, err := fetcher.NewQuota(f, 5, 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestQuota_Fetch_WithinLimit(t *testing.T) {
	stub := &stubFetcher{result: map[string]interface{}{"ok": true}}
	q, err := fetcher.NewQuota(stub, 3, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := q.Fetch("svc", "http://example.com")
		if err != nil {
			t.Fatalf("fetch %d failed: %v", i, err)
		}
	}
	if got := q.Remaining(); got != 0 {
		t.Errorf("expected 0 remaining, got %d", got)
	}
}

func TestQuota_Fetch_ExceedsLimit(t *testing.T) {
	stub := &stubFetcher{result: map[string]interface{}{"ok": true}}
	q, _ := fetcher.NewQuota(stub, 2, time.Minute)
	q.Fetch("svc", "http://example.com") //nolint
	q.Fetch("svc", "http://example.com") //nolint
	_, err := q.Fetch("svc", "http://example.com")
	if !errors.Is(err, fetcher.ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
}

func TestQuota_WindowReset(t *testing.T) {
	stub := &stubFetcher{result: map[string]interface{}{"ok": true}}
	q, _ := fetcher.NewQuota(stub, 1, 50*time.Millisecond)
	_, err := q.Fetch("svc", "http://example.com")
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	_, err = q.Fetch("svc", "http://example.com")
	if !errors.Is(err, fetcher.ErrQuotaExceeded) {
		t.Fatalf("expected quota exceeded, got %v", err)
	}
	time.Sleep(60 * time.Millisecond)
	_, err = q.Fetch("svc", "http://example.com")
	if err != nil {
		t.Fatalf("fetch after window reset failed: %v", err)
	}
}

func TestQuota_Concurrent_DoesNotExceedMax(t *testing.T) {
	stub := &stubFetcher{result: map[string]interface{}{"ok": true}}
	const max = 10
	q, _ := fetcher.NewQuota(stub, max, time.Minute)
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		success int
	)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := q.Fetch("svc", "http://example.com")
			if err == nil {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if success != max {
		t.Errorf("expected exactly %d successes, got %d", max, success)
	}
}

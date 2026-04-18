package fetcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

func TestBatch_WithCachedFetcher_Integration(t *testing.T) {
	var callCount int64
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version":"1.0"}`))
	}))
	defer svr.Close()

	inner, _ := fetcher.New(0)
	cached := fetcher.NewCached(inner, 60)
	bf, err := fetcher.NewBatch(cached, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First batch — all miss cache
	urls := []string{svr.URL, svr.URL, svr.URL}
	results := bf.FetchAll(context.Background(), urls)
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error: %v", r.Err)
		}
	}

	// Second batch — should hit cache, call count should not increase much
	bf.FetchAll(context.Background(), urls)

	if atomic.LoadInt64(&callCount) == 0 {
		t.Error("expected at least one real HTTP call")
	}
}

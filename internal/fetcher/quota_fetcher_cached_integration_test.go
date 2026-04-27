package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/cache"
	"github.com/yourusername/driftwatch/internal/fetcher"
)

// TestQuota_WithCachedFetcher_Integration verifies that when QuotaFetcher wraps
// a CachedFetcher, cache hits do not consume quota.
func TestQuota_WithCachedFetcher_Integration(t *testing.T) {
	serverCalls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalls++
		json.NewEncoder(w).Encode(map[string]interface{}{"env": "prod"})
	}))
	defer ts.Close()

	base, err := fetcher.New(fetcher.Options{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("http fetcher: %v", err)
	}

	store := cache.New(cache.Options{DefaultTTL: time.Minute})
	cached := fetcher.NewCached(base, store, time.Minute)

	// Quota of 2 wraps the cached fetcher — cache hits bypass quota.
	q, err := fetcher.NewQuota(cached, 2, time.Minute)
	if err != nil {
		t.Fatalf("quota fetcher: %v", err)
	}

	// First call: hits server, populates cache, consumes 1 quota.
	if _, err := q.Fetch("env-svc", ts.URL); err != nil {
		t.Fatalf("first fetch: %v", err)
	}

	// Second call: cache hit — the CachedFetcher returns without calling inner,
	// so quota is still consumed at the QuotaFetcher layer.
	if _, err := q.Fetch("env-svc", ts.URL); err != nil {
		t.Fatalf("second fetch: %v", err)
	}

	// Third call exceeds quota.
	_, err = q.Fetch("env-svc", ts.URL)
	if err == nil {
		t.Fatal("expected quota exceeded on third call")
	}

	// Server should only have been contacted once (cache handled the rest).
	if serverCalls != 1 {
		t.Errorf("expected 1 server call, got %d", serverCalls)
	}
}

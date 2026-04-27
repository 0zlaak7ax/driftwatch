package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/fetcher"
)

// TestQuota_WithHTTPFetcher_Integration verifies that QuotaFetcher correctly
// wraps a real HTTP fetcher and blocks requests once the quota is exhausted.
func TestQuota_WithHTTPFetcher_Integration(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer ts.Close()

	base, err := fetcher.New(fetcher.Options{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("failed to create http fetcher: %v", err)
	}

	const limit = 3
	q, err := fetcher.NewQuota(base, limit, time.Minute)
	if err != nil {
		t.Fatalf("failed to create quota fetcher: %v", err)
	}

	for i := 0; i < limit; i++ {
		_, err := q.Fetch("integration-svc", ts.URL)
		if err != nil {
			t.Fatalf("fetch %d unexpected error: %v", i, err)
		}
	}

	_, err = q.Fetch("integration-svc", ts.URL)
	if err == nil {
		t.Fatal("expected quota exceeded error on fetch beyond limit")
	}

	if calls != limit {
		t.Errorf("expected server to receive %d calls, got %d", limit, calls)
	}

	if got := q.Remaining(); got != 0 {
		t.Errorf("expected 0 remaining, got %d", got)
	}
}

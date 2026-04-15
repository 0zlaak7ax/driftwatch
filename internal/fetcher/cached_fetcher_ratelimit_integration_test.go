package fetcher_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/cache"
	"github.com/example/driftwatch/internal/fetcher"
	"github.com/example/driftwatch/internal/ratelimit"
)

func TestCachedRateLimited_Integration(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer ts.Close()

	base := fetcher.New(fetcher.Options{Timeout: 2 * time.Second})

	c, err := cache.New(cache.Options{DefaultTTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("cache.New: %v", err)
	}
	cached := fetcher.NewCached(base, c)

	l, err := ratelimit.New(ratelimit.Config{Rate: 10, Interval: time.Second})
	if err != nil {
		t.Fatalf("ratelimit.New: %v", err)
	}
	rl, err := fetcher.NewRateLimited(cached, l)
	if err != nil {
		t.Fatalf("NewRateLimited: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		res, err := rl.Fetch(ctx, ts.URL)
		if err != nil {
			t.Fatalf("Fetch %d: %v", i, err)
		}
		if res["status"] != "ok" {
			t.Errorf("unexpected result on call %d: %v", i, res)
		}
	}

	// Cache should mean only 1 real HTTP call despite 3 fetches.
	if calls != 1 {
		t.Errorf("expected 1 upstream call due to caching, got %d", calls)
	}
}

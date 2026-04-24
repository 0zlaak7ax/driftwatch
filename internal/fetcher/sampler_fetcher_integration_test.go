package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/driftwatch/internal/fetcher"
)

// TestSampler_WithHTTPFetcher_Integration verifies that SamplerFetcher correctly
// delegates to a real HTTPFetcher and caches the result for subsequent skipped calls.
func TestSampler_WithHTTPFetcher_Integration(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"version": "1.2.3"})
	}))
	defer ts.Close()

	http := fetcher.New(0)
	sampler, err := fetcher.NewSampler(http, 1.0)
	if err != nil {
		t.Fatalf("NewSampler error: %v", err)
	}

	// First call — must hit the server.
	res, err := sampler.Fetch(ts.URL)
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if res["version"] != "1.2.3" {
		t.Fatalf("unexpected version: %v", res["version"])
	}
	if callCount != 1 {
		t.Fatalf("expected 1 server call, got %d", callCount)
	}

	// With rate=1.0 every call hits the server; confirm consistent results.
	res2, err := sampler.Fetch(ts.URL)
	if err != nil {
		t.Fatalf("second Fetch error: %v", err)
	}
	if res2["version"] != "1.2.3" {
		t.Fatalf("unexpected version on second call: %v", res2["version"])
	}
	if callCount != 2 {
		t.Fatalf("expected 2 server calls, got %d", callCount)
	}
}

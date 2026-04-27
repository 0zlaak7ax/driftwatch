package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/driftwatch/internal/fetcher"
)

func TestPriority_WithHTTPFetcher_Integration(t *testing.T) {
	// Primary server returns valid JSON.
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"source": "primary", "version": "1.0"})
	}))
	defer primary.Close()

	// Secondary server also returns valid JSON (should not be reached).
	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"source": "secondary", "version": "2.0"})
	}))
	defer secondary.Close()

	httpFetcher := fetcher.New(0)

	f, err := fetcher.NewPriority([]fetcher.PriorityEntry{
		{Fetcher: httpFetcher, Priority: 1},
		{Fetcher: httpFetcher, Priority: 5},
	})
	if err != nil {
		t.Fatalf("NewPriority error: %v", err)
	}

	res, err := f.Fetch(primary.URL)
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if res["source"] != "primary" {
		t.Errorf("expected source=primary, got %v", res["source"])
	}
}

func TestPriority_PrimaryDown_UsesSecondary_Integration(t *testing.T) {
	// Secondary server returns valid JSON.
	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"source": "secondary"})
	}))
	defer secondary.Close()

	httpFetcher := fetcher.New(0)

	// Use a stub that always fails as the high-priority entry.
	failing := &stubPriority{err: http.ErrHandlerTimeout}

	f, err := fetcher.NewPriority([]fetcher.PriorityEntry{
		{Fetcher: failing, Priority: 1},
		{Fetcher: httpFetcher, Priority: 2},
	})
	if err != nil {
		t.Fatalf("NewPriority error: %v", err)
	}

	res, err := f.Fetch(secondary.URL)
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if res["source"] != "secondary" {
		t.Errorf("expected source=secondary, got %v", res["source"])
	}
}

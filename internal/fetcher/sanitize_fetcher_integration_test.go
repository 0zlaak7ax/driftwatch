package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

func TestSanitize_WithHTTPFetcher_Integration(t *testing.T) {
	payload := map[string]interface{}{
		"env":     "  PRODUCTION  ",
		"version": "  1.4.2  ",
		"region":  "EU-WEST-1",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	base, err := fetcher.New(0)
	if err != nil {
		t.Fatalf("failed to create HTTP fetcher: %v", err)
	}

	sanitized, err := fetcher.NewSanitize(base,
		fetcher.TrimSpaceRule(),
		fetcher.LowercaseRule(),
	)
	if err != nil {
		t.Fatalf("failed to create sanitize fetcher: %v", err)
	}

	result, err := sanitized.Fetch(ts.URL)
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}

	cases := map[string]string{
		"env":     "production",
		"version": "1.4.2",
		"region":  "eu-west-1",
	}
	for key, want := range cases {
		got, ok := result[key].(string)
		if !ok {
			t.Errorf("key %q: expected string value", key)
			continue
		}
		if got != want {
			t.Errorf("key %q: expected %q, got %q", key, want, got)
		}
	}
}

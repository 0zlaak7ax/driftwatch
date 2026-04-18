package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

// TestAuth_WithHTTPFetcher_Integration verifies that NewAuth used as a
// transport correctly injects the Authorization header when combined with
// the real HTTP fetcher via a custom http.Client.
func TestAuth_WithHTTPFetcher_Integration(t *testing.T) {
	payload := map[string]interface{}{"env": "production", "replicas": float64(3)}

	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	stub := &stubFetcher{result: payload}
	af, err := fetcher.NewAuth(stub, "Bearer", "integration-token")
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}

	client := &http.Client{Transport: af}
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if receivedAuth != "Bearer integration-token" {
		t.Errorf("auth header mismatch: got %q", receivedAuth)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["env"] != "production" {
		t.Errorf("unexpected body: %v", body)
	}
}

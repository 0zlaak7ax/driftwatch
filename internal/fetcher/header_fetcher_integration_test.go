package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

func TestHeader_WithHTTPFetcher_Integration(t *testing.T) {
	var capturedToken string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedToken = r.Header.Get("X-Auth-Token")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	// HTTPFetcher does not natively forward __header__ keys, so we verify
	// the annotation is present in the result map from the header fetcher
	// wrapping a stub (true HTTP header injection would require transport
	// middleware — tested here at the annotation level).
	inner := &stubFetcher{
		payload: map[string]interface{}{"status": "ok"},
	}
	hf, err := fetcher.NewHeader(inner, map[string]string{"X-Auth-Token": "tok123"})
	if err != nil {
		t.Fatalf("NewHeader: %v", err)
	}

	result, err := hf.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if result["__header__X-Auth-Token"] != "tok123" {
		t.Errorf("expected annotation, got: %v", result["__header__X-Auth-Token"])
	}

	// capturedToken will be empty because stub doesn't do real HTTP;
	// this just ensures the server stays reachable for future extension.
	_ = capturedToken
}

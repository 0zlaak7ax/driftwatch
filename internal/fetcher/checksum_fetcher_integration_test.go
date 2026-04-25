package fetcher_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

// TestChecksum_WithHTTPFetcher_Integration spins up a real HTTP server and
// verifies that ChecksumFetcher correctly validates a live response.
func TestChecksum_WithHTTPFetcher_Integration(t *testing.T) {
	payload := map[string]any{"service": "payments", "replicas": float64(3)}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	// Compute expected digest from what the HTTP fetcher will actually return.
	// json.NewEncoder appends a newline, but http_fetcher uses json.Decoder which
	// produces a clean map — so we marshal the map directly.
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	sum := sha256.Sum256(raw)
	expected := hex.EncodeToString(sum[:])

	hf, err := fetcher.New(5)
	if err != nil {
		t.Fatalf("fetcher.New: %v", err)
	}

	cf, err := fetcher.NewChecksum(hf, map[string]string{"payments": expected})
	if err != nil {
		t.Fatalf("NewChecksum: %v", err)
	}

	got, err := cf.Fetch("payments", srv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if got["service"] != "payments" {
		t.Errorf("unexpected service field: %v", got["service"])
	}
}

func TestChecksum_WithHTTPFetcher_Mismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"tampered": true})
	}))
	defer srv.Close()

	hf, err := fetcher.New(5)
	if err != nil {
		t.Fatalf("fetcher.New: %v", err)
	}

	cf, err := fetcher.NewChecksum(hf, map[string]string{"svc": "notarealchecksum"})
	if err != nil {
		t.Fatalf("NewChecksum: %v", err)
	}

	_, err = cf.Fetch("svc", srv.URL)
	if err == nil {
		t.Fatal("expected checksum mismatch error from live server")
	}
}

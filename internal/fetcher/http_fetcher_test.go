package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/fetcher"
)

func TestFetch_OK(t *testing.T) {
	payload := map[string]interface{}{"version": "1.2.3", "replicas": float64(3)}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	f := fetcher.New(5 * time.Second)
	got, err := f.Fetch(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["version"] != "1.2.3" {
		t.Errorf("version: got %v, want 1.2.3", got["version"])
	}
	if got["replicas"] != float64(3) {
		t.Errorf("replicas: got %v, want 3", got["replicas"])
	}
}

func TestFetch_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	f := fetcher.New(5 * time.Second)
	_, err := f.Fetch(ts.URL)
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

func TestFetch_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	f := fetcher.New(5 * time.Second)
	_, err := f.Fetch(ts.URL)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestFetch_Unreachable(t *testing.T) {
	f := fetcher.New(500 * time.Millisecond)
	_, err := f.Fetch("http://127.0.0.1:19999/nope")
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}

func TestNew_DefaultTimeout(t *testing.T) {
	// Passing 0 should fall back to the 10 s default without panicking.
	f := fetcher.New(0)
	if f == nil {
		t.Fatal("expected non-nil fetcher")
	}
}

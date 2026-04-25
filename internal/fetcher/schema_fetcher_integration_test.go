package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

func TestSchema_WithHTTPFetcher_Integration(t *testing.T) {
	payload := map[string]any{
		"version":  "3.1.0",
		"replicas": float64(2),
		"healthy":  true,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	base, err := fetcher.New(0)
	if err != nil {
		t.Fatalf("failed to create http fetcher: %v", err)
	}

	rules := []fetcher.SchemaRule{
		{Field: "version", Required: true, Type: "string"},
		{Field: "replicas", Required: true, Type: "number"},
		{Field: "healthy", Required: false, Type: "bool"},
	}
	f, err := fetcher.NewSchema(base, rules)
	if err != nil {
		t.Fatalf("failed to create schema fetcher: %v", err)
	}

	data, err := f.Fetch("integration-svc", ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["version"] != "3.1.0" {
		t.Errorf("expected version 3.1.0, got %v", data["version"])
	}
}

func TestSchema_WithHTTPFetcher_TypeMismatch(t *testing.T) {
	payload := map[string]any{
		"version": 999,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	base, _ := fetcher.New(0)
	rules := []fetcher.SchemaRule{
		{Field: "version", Required: true, Type: "string"},
	}
	f, _ := fetcher.NewSchema(base, rules)

	_, err := f.Fetch("integration-svc", ts.URL)
	if err == nil {
		t.Fatal("expected type mismatch error from integration test")
	}
}

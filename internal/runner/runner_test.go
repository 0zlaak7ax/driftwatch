package runner_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/runner"
)

func makeServer(t *testing.T, body map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
}

func TestRun_NoDrift(t *testing.T) {
	srv := makeServer(t, map[string]string{"version": "1.2.3", "region": "us-east-1"})
	defer srv.Close()

	cfg := &config.Config{
		Services: []config.Service{
			{
				Name: "api",
				URL:  srv.URL,
				Expected: map[string]string{
					"version": "1.2.3",
					"region":  "us-east-1",
				},
			},
		},
	}

	r, err := runner.New(cfg, "text")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	drifted, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if drifted {
		t.Error("expected no drift, got drift=true")
	}
}

func TestRun_WithDrift(t *testing.T) {
	srv := makeServer(t, map[string]string{"version": "2.0.0"})
	defer srv.Close()

	cfg := &config.Config{
		Services: []config.Service{
			{
				Name:     "worker",
				URL:      srv.URL,
				Expected: map[string]string{"version": "1.0.0"},
			},
		},
	}

	r, err := runner.New(cfg, "json")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	drifted, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !drifted {
		t.Error("expected drift=true, got false")
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	cfg := &config.Config{}
	_, err := runner.New(cfg, "xml")
	if err == nil {
		t.Error("expected error for invalid format, got nil")
	}
}

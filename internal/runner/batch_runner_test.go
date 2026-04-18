package runner_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/fetcher"
	"github.com/driftwatch/internal/runner"
)

func makeBatchConfig(urls ...string) *config.Config {
	svcs := make([]config.Service, len(urls))
	for i, u := range urls {
		svcs[i] = config.Service{Name: fmt.Sprintf("svc%d", i), URL: u}
	}
	return &config.Config{Services: svcs}
}

func TestNewBatchRunner_NilConfig(t *testing.T) {
	f, _ := fetcher.New(0)
	_, err := runner.NewBatchRunner(nil, f, 2)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestNewBatchRunner_InvalidWorkers(t *testing.T) {
	f, _ := fetcher.New(0)
	cfg := &config.Config{}
	_, err := runner.NewBatchRunner(cfg, f, 0)
	if err == nil {
		t.Fatal("expected error for 0 workers")
	}
}

func TestBatchRunner_Prefetch_AllSuccess(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer svr.Close()

	f, _ := fetcher.New(0)
	cfg := makeBatchConfig(svr.URL, svr.URL)
	br, err := runner.NewBatchRunner(cfg, f, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errs := br.Prefetch(context.Background())
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestBatchRunner_Prefetch_PartialFailure(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer svr.Close()

	f, _ := fetcher.New(0)
	cfg := makeBatchConfig(svr.URL, "http://127.0.0.1:1")
	br, _ := runner.NewBatchRunner(cfg, f, 2)

	errs := br.Prefetch(context.Background())
	if len(errs) == 0 {
		t.Error("expected at least one prefetch error")
	}
}

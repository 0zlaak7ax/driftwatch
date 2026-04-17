package fetcher_test

import (
	"errors"
	"testing"

	"github.com/driftwatch/internal/fetcher"
	"github.com/driftwatch/internal/metrics"
)

type stubFetcherMF struct {
	result map[string]any
	err    error
}

func (s *stubFetcherMF) Fetch(_ string) (map[string]any, error) {
	return s.result, s.err
}

func TestNewMetrics_NilInner(t *testing.T) {
	_, err := fetcher.NewMetrics(nil, metrics.New(), "svc")
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewMetrics_NilStore(t *testing.T) {
	_, err := fetcher.NewMetrics(&stubFetcherMF{}, nil, "svc")
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNewMetrics_EmptyService(t *testing.T) {
	_, err := fetcher.NewMetrics(&stubFetcherMF{}, metrics.New(), "")
	if err == nil {
		t.Fatal("expected error for empty service")
	}
}

func TestMetrics_Fetch_Success(t *testing.T) {
	store := metrics.New()
	stub := &stubFetcherMF{result: map[string]any{"k": "v"}}
	mf, err := fetcher.NewMetrics(stub, store, "svc-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, err := mf.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if res["k"] != "v" {
		t.Errorf("unexpected result: %v", res)
	}
	runs := store.All()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if !runs[0].Success {
		t.Error("expected success=true")
	}
	if runs[0].Service != "svc-a" {
		t.Errorf("expected service svc-a, got %s", runs[0].Service)
	}
}

func TestMetrics_Fetch_Error(t *testing.T) {
	store := metrics.New()
	stub := &stubFetcherMF{err: errors.New("boom")}
	mf, _ := fetcher.NewMetrics(stub, store, "svc-b")
	_, err := mf.Fetch("http://example.com")
	if err == nil {
		t.Fatal("expected error")
	}
	runs := store.All()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Success {
		t.Error("expected success=false")
	}
}

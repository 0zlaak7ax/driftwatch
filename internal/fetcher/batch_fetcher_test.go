package fetcher_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

func TestNewBatch_NilInner(t *testing.T) {
	_, err := fetcher.NewBatch(nil, 2)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewBatch_InvalidWorkers(t *testing.T) {
	f, _ := fetcher.New(0)
	_, err := fetcher.NewBatch(f, 0)
	if err == nil {
		t.Fatal("expected error for 0 workers")
	}
}

func TestBatch_FetchAll_Success(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer svr.Close()

	inner, _ := fetcher.New(0)
	bf, _ := fetcher.NewBatch(inner, 2)

	urls := []string{svr.URL, svr.URL, svr.URL}
	results := bf.FetchAll(context.Background(), urls)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error: %v", r.Err)
		}
		if r.Data["status"] != "ok" {
			t.Errorf("unexpected data: %v", r.Data)
		}
	}
}

func TestBatch_FetchAll_PartialError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer svr.Close()

	inner, _ := fetcher.New(0)
	bf, _ := fetcher.NewBatch(inner, 2)

	urls := []string{svr.URL, "http://127.0.0.1:1"}
	results := bf.FetchAll(context.Background(), urls)

	if len(results) != 2 {
		t.Fatalf("expected 2 results")
	}
	var errs []error
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, r.Err)
		}
	}
	if len(errs) == 0 {
		t.Error("expected at least one error")
	}
}

func TestBatch_FetchAll_Empty(t *testing.T) {
	inner, _ := fetcher.New(0)
	bf, _ := fetcher.NewBatch(inner, 2)
	results := bf.FetchAll(context.Background(), []string{})
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
	_ = errors.New("unused")
}

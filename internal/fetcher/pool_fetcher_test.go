package fetcher_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type countingFetcher struct {
	calls atomic.Int32
	errOn string
}

func (c *countingFetcher) Fetch(_ context.Context, url string) (map[string]interface{}, error) {
	c.calls.Add(1)
	if url == c.errOn {
		return nil, errors.New("fetch error")
	}
	return map[string]interface{}{"url": url}, nil
}

func TestNewPool_NilInner(t *testing.T) {
	_, err := fetcher.NewPool(nil, 2)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewPool_InvalidWorkers(t *testing.T) {
	_, err := fetcher.NewPool(&countingFetcher{}, 0)
	if err == nil {
		t.Fatal("expected error for zero workers")
	}
}

func TestPool_FetchAll_Success(t *testing.T) {
	cf := &countingFetcher{}
	p, err := fetcher.NewPool(cf, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	urls := []string{"http://a", "http://b", "http://c"}
	results := p.FetchAll(context.Background(), urls)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result %d unexpected error: %v", i, r.Err)
		}
		if r.URL != urls[i] {
			t.Errorf("result %d URL mismatch: got %s want %s", i, r.URL, urls[i])
		}
	}
	if int(cf.calls.Load()) != 3 {
		t.Errorf("expected 3 calls, got %d", cf.calls.Load())
	}
}

func TestPool_FetchAll_PartialError(t *testing.T) {
	cf := &countingFetcher{errOn: "http://b"}
	p, _ := fetcher.NewPool(cf, 2)
	results := p.FetchAll(context.Background(), []string{"http://a", "http://b"})
	if results[0].Err != nil {
		t.Errorf("expected no error for http://a")
	}
	if results[1].Err == nil {
		t.Errorf("expected error for http://b")
	}
}

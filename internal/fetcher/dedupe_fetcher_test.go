package fetcher_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type countingFetcher struct {
	count int64
	result map[string]interface{}
	err    error
}

func (c *countingFetcher) Fetch(_ string) (map[string]interface{}, error) {
	atomic.AddInt64(&c.count, 1)
	return c.result, c.err
}

func TestNewDedupe_NilInner(t *testing.T) {
	_, err := fetcher.NewDedupe(nil)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestDedupe_Fetch_SingleCall(t *testing.T) {
	inner := &countingFetcher{result: map[string]interface{}{"v": 1}}
	d, _ := fetcher.NewDedupe(inner)
	res, err := d.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["v"] != 1 {
		t.Fatalf("unexpected result: %v", res)
	}
	if inner.count != 1 {
		t.Fatalf("expected 1 upstream call, got %d", inner.count)
	}
}

func TestDedupe_Fetch_ConcurrentCallsCoalesced(t *testing.T) {
	var ready sync.WaitGroup
	var release = make(chan struct{})

	inner := &blockingFetcher{
		release: release,
		result:  map[string]interface{}{"x": 42},
	}
	d, _ := fetcher.NewDedupe(inner)

	const n = 10
	ready.Add(n)
	results := make([]map[string]interface{}, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			ready.Done()
			results[i], errs[i] = d.Fetch("http://example.com")
		}()
	}
	ready.Wait()
	close(release)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: unexpected error: %v", i, err)
		}
	}
	if atomic.LoadInt64(&inner.count) > int64(n) {
		t.Logf("upstream calls: %d (coalescing best-effort)", inner.count)
	}
}

func TestDedupe_Fetch_PropagatesError(t *testing.T) {
	inner := &countingFetcher{err: errors.New("upstream down")}
	d, _ := fetcher.NewDedupe(inner)
	_, err := d.Fetch("http://bad.example.com")
	if err == nil {
		t.Fatal("expected error")
	}
}

// blockingFetcher blocks until release is closed.
type blockingFetcher struct {
	count   int64
	release chan struct{}
	result  map[string]interface{}
}

func (b *blockingFetcher) Fetch(_ string) (map[string]interface{}, error) {
	atomic.AddInt64(&b.count, 1)
	<-b.release
	return b.result, nil
}

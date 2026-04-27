package fetcher_test

import (
	"errors"
	"testing"

	"github.com/example/driftwatch/internal/fetcher"
)

type stubPriority struct {
	result map[string]interface{}
	err    error
	calls  int
}

func (s *stubPriority) Fetch(_ string) (map[string]interface{}, error) {
	s.calls++
	return s.result, s.err
}

func TestNewPriority_Empty_ReturnsError(t *testing.T) {
	_, err := fetcher.NewPriority(nil)
	if err == nil {
		t.Fatal("expected error for empty entries")
	}
}

func TestNewPriority_NilFetcher_ReturnsError(t *testing.T) {
	_, err := fetcher.NewPriority([]fetcher.PriorityEntry{{Fetcher: nil, Priority: 1}})
	if err == nil {
		t.Fatal("expected error for nil fetcher")
	}
}

func TestPriority_FirstSucceeds(t *testing.T) {
	high := &stubPriority{result: map[string]interface{}{"src": "high"}, err: nil}
	low := &stubPriority{result: map[string]interface{}{"src": "low"}, err: nil}

	f, err := fetcher.NewPriority([]fetcher.PriorityEntry{
		{Fetcher: low, Priority: 10},
		{Fetcher: high, Priority: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, err := f.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if res["src"] != "high" {
		t.Errorf("expected high-priority result, got %v", res["src"])
	}
	if high.calls != 1 {
		t.Errorf("expected high-priority fetcher called once, got %d", high.calls)
	}
	if low.calls != 0 {
		t.Errorf("expected low-priority fetcher not called, got %d", low.calls)
	}
}

func TestPriority_FirstFails_FallsToSecond(t *testing.T) {
	failing := &stubPriority{err: errors.New("unavailable")}
	working := &stubPriority{result: map[string]interface{}{"src": "fallback"}}

	f, err := fetcher.NewPriority([]fetcher.PriorityEntry{
		{Fetcher: failing, Priority: 1},
		{Fetcher: working, Priority: 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, err := f.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if res["src"] != "fallback" {
		t.Errorf("expected fallback result, got %v", res["src"])
	}
}

func TestPriority_AllFail_ReturnsAggregateError(t *testing.T) {
	a := &stubPriority{err: errors.New("err-a")}
	b := &stubPriority{err: errors.New("err-b")}

	f, err := fetcher.NewPriority([]fetcher.PriorityEntry{
		{Fetcher: a, Priority: 1},
		{Fetcher: b, Priority: 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, fetchErr := f.Fetch("http://example.com")
	if fetchErr == nil {
		t.Fatal("expected error when all fetchers fail")
	}
	if a.calls != 1 || b.calls != 1 {
		t.Errorf("expected both fetchers called, got a=%d b=%d", a.calls, b.calls)
	}
}

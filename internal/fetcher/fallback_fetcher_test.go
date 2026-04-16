package fetcher_test

import (
	"context"
	"errors"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type stubFetcher struct {
	data map[string]interface{}
	err  error
}

func (s *stubFetcher) Fetch(_ context.Context, _ string) (map[string]interface{}, error) {
	return s.data, s.err
}

func TestNewFallback_NilPrimary(t *testing.T) {
	_, err := fetcher.NewFallback(nil, &stubFetcher{})
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNewFallback_NilFallback(t *testing.T) {
	_, err := fetcher.NewFallback(&stubFetcher{}, nil)
	if err == nil {
		t.Fatal("expected error for nil fallback")
	}
}

func TestFallback_PrimarySucceeds(t *testing.T) {
	expected := map[string]interface{}{"version": "1.0"}
	primary := &stubFetcher{data: expected}
	fallback := &stubFetcher{data: map[string]interface{}{"version": "0.0"}}

	f, err := fetcher.NewFallback(primary, fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := f.Fetch(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["version"] != "1.0" {
		t.Errorf("expected primary result, got %v", got)
	}
}

func TestFallback_PrimaryFails_UsesFallback(t *testing.T) {
	expected := map[string]interface{}{"version": "2.0"}
	primary := &stubFetcher{err: errors.New("primary down")}
	fallback := &stubFetcher{data: expected}

	f, err := fetcher.NewFallback(primary, fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := f.Fetch(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["version"] != "2.0" {
		t.Errorf("expected fallback result, got %v", got)
	}
}

func TestFallback_BothFail_ReturnsError(t *testing.T) {
	primary := &stubFetcher{err: errors.New("primary down")}
	fallback := &stubFetcher{err: errors.New("fallback down")}

	f, err := fetcher.NewFallback(primary, fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = f.Fetch(context.Background(), "http://example.com")
	if err == nil {
		t.Fatal("expected error when both fetchers fail")
	}
}

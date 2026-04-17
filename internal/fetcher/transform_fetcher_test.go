package fetcher_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type stubFetcher struct {
	data map[string]string
	err  error
}

func (s *stubFetcher) Fetch(_ string) (map[string]string, error) {
	return s.data, s.err
}

func TestNewTransform_NilInner(t *testing.T) {
	_, err := fetcher.NewTransform(nil, func(m map[string]string) (map[string]string, error) { return m, nil })
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewTransform_NilFunc(t *testing.T) {
	_, err := fetcher.NewTransform(&stubFetcher{}, nil)
	if err == nil {
		t.Fatal("expected error for nil func")
	}
}

func TestTransform_Fetch_AppliesFunc(t *testing.T) {
	inner := &stubFetcher{data: map[string]string{"key": "hello"}}
	upperFn := func(m map[string]string) (map[string]string, error) {
		out := make(map[string]string, len(m))
		for k, v := range m {
			out[k] = strings.ToUpper(v)
		}
		return out, nil
	}
	f, err := fetcher.NewTransform(inner, upperFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := f.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if result["key"] != "HELLO" {
		t.Errorf("expected HELLO, got %s", result["key"])
	}
}

func TestTransform_Fetch_InnerError(t *testing.T) {
	inner := &stubFetcher{err: errors.New("fetch failed")}
	f, _ := fetcher.NewTransform(inner, func(m map[string]string) (map[string]string, error) { return m, nil })
	_, err := f.Fetch("http://example.com")
	if err == nil {
		t.Fatal("expected error from inner fetcher")
	}
}

func TestTransform_Fetch_TransformError(t *testing.T) {
	inner := &stubFetcher{data: map[string]string{"k": "v"}}
	badFn := func(m map[string]string) (map[string]string, error) {
		return nil, errors.New("transform error")
	}
	f, _ := fetcher.NewTransform(inner, badFn)
	_, err := f.Fetch("http://example.com")
	if err == nil {
		t.Fatal("expected error from transform func")
	}
	if !strings.Contains(err.Error(), "transform failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

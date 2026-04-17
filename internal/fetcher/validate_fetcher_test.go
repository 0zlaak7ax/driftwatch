package fetcher_test

import (
	"errors"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type stubValidateFetcher struct {
	data map[string]interface{}
	err  error
}

func (s *stubValidateFetcher) Fetch(_, _ string) (map[string]interface{}, error) {
	return s.data, s.err
}

func TestNewValidate_NilInner(t *testing.T) {
	_, err := fetcher.NewValidate(nil, func(_ string, _ map[string]interface{}) error { return nil })
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewValidate_NilFunc(t *testing.T) {
	inner := &stubValidateFetcher{data: map[string]interface{}{}}
	_, err := fetcher.NewValidate(inner, nil)
	if err == nil {
		t.Fatal("expected error for nil func")
	}
}

func TestValidate_Fetch_PassesValidation(t *testing.T) {
	inner := &stubValidateFetcher{data: map[string]interface{}{"version": "1.2.3"}}
	f, err := fetcher.NewValidate(inner, func(_ string, data map[string]interface{}) error {
		if _, ok := data["version"]; !ok {
			return errors.New("missing version")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := f.Fetch("svc", "http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if got["version"] != "1.2.3" {
		t.Errorf("unexpected data: %v", got)
	}
}

func TestValidate_Fetch_FailsValidation(t *testing.T) {
	inner := &stubValidateFetcher{data: map[string]interface{}{"status": "ok"}}
	f, _ := fetcher.NewValidate(inner, func(_ string, data map[string]interface{}) error {
		if _, ok := data["version"]; !ok {
			return errors.New("missing version")
		}
		return nil
	})
	_, err := f.Fetch("svc", "http://example.com")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidate_Fetch_InnerError(t *testing.T) {
	inner := &stubValidateFetcher{err: errors.New("network error")}
	f, _ := fetcher.NewValidate(inner, func(_ string, _ map[string]interface{}) error { return nil })
	_, err := f.Fetch("svc", "http://example.com")
	if err == nil {
		t.Fatal("expected inner error to propagate")
	}
}

package fetcher_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type stubSanitize struct {
	data map[string]interface{}
	err  error
}

func (s *stubSanitize) Fetch(_ string) (map[string]interface{}, error) {
	return s.data, s.err
}

func TestNewSanitize_NilInner(t *testing.T) {
	_, err := fetcher.NewSanitize(nil, fetcher.TrimSpaceRule())
	if err == nil {
		t.Fatal("expected error for nil inner fetcher")
	}
}

func TestNewSanitize_NoRules(t *testing.T) {
	inner := &stubSanitize{data: map[string]interface{}{}}
	_, err := fetcher.NewSanitize(inner)
	if err == nil {
		t.Fatal("expected error when no rules provided")
	}
}

func TestSanitize_Fetch_TrimSpace(t *testing.T) {
	inner := &stubSanitize{
		data: map[string]interface{}{"env": "  production  ", "version": "  1.2.3"},
	}
	f, err := fetcher.NewSanitize(inner, fetcher.TrimSpaceRule())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := f.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if got := result["env"]; got != "production" {
		t.Errorf("expected 'production', got %q", got)
	}
	if got := result["version"]; got != "1.2.3" {
		t.Errorf("expected '1.2.3', got %q", got)
	}
}

func TestSanitize_Fetch_Lowercase(t *testing.T) {
	inner := &stubSanitize{
		data: map[string]interface{}{"region": "US-EAST-1"},
	}
	f, err := fetcher.NewSanitize(inner, fetcher.LowercaseRule())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := f.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if got := result["region"]; got != "us-east-1" {
		t.Errorf("expected 'us-east-1', got %q", got)
	}
}

func TestSanitize_Fetch_MaxLength(t *testing.T) {
	inner := &stubSanitize{
		data: map[string]interface{}{"desc": "this is a very long description string"},
	}
	f, err := fetcher.NewSanitize(inner, fetcher.MaxLengthRule(10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := f.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	got, _ := result["desc"].(string)
	if len(got) > 10 {
		t.Errorf("expected value truncated to 10 chars, got len=%d", len(got))
	}
}

func TestSanitize_Fetch_InnerError(t *testing.T) {
	inner := &stubSanitize{err: errors.New("network error")}
	f, err := fetcher.NewSanitize(inner, fetcher.TrimSpaceRule())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = f.Fetch("http://example.com")
	if err == nil || !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected inner error to propagate, got: %v", err)
	}
}

func TestSanitize_Fetch_RuleError(t *testing.T) {
	inner := &stubSanitize{
		data: map[string]interface{}{"key": "value"},
	}
	badRule := func(key, value string) (string, error) {
		return "", errors.New("rule failed")
	}
	f, err := fetcher.NewSanitize(inner, badRule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = f.Fetch("http://example.com")
	if err == nil || !strings.Contains(err.Error(), "rule failed") {
		t.Errorf("expected rule error to propagate, got: %v", err)
	}
}

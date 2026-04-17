package fetcher_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"driftwatch/internal/fetcher"
)

type mockFetcher struct {
	result map[string]interface{}
	err    error
}

func (m *mockFetcher) Fetch(_ string) (map[string]interface{}, error) {
	return m.result, m.err
}

func TestNewLogging_NilInner(t *testing.T) {
	_, err := fetcher.NewLogging(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil inner fetcher")
	}
}

func TestLogging_Fetch_Success(t *testing.T) {
	inner := &mockFetcher{result: map[string]interface{}{"version": "1.0", "env": "prod"}}
	var buf bytes.Buffer
	f, err := fetcher.NewLogging(inner, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := f.Fetch("http://example.com/status")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 fields, got %d", len(result))
	}

	log := buf.String()
	if !strings.Contains(log, "fetch ok") {
		t.Errorf("expected 'fetch ok' in log, got: %s", log)
	}
	if !strings.Contains(log, "fields=2") {
		t.Errorf("expected 'fields=2' in log, got: %s", log)
	}
}

func TestLogging_Fetch_Error(t *testing.T) {
	inner := &mockFetcher{err: errors.New("connection refused")}
	var buf bytes.Buffer
	f, err := fetcher.NewLogging(inner, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, fetchErr := f.Fetch("http://example.com/status")
	if fetchErr == nil {
		t.Fatal("expected fetch error")
	}

	log := buf.String()
	if !strings.Contains(log, "fetch error") {
		t.Errorf("expected 'fetch error' in log, got: %s", log)
	}
	if !strings.Contains(log, "connection refused") {
		t.Errorf("expected error message in log, got: %s", log)
	}
}

func TestNewLogging_NilWriter_UsesStderr(t *testing.T) {
	inner := &mockFetcher{result: map[string]interface{}{}}
	f, err := fetcher.NewLogging(inner, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil fetcher")
	}
}

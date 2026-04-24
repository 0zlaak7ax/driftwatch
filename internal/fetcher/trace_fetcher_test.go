package fetcher_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type traceMockFetcher struct {
	result map[string]interface{}
	err    error
}

func (m *traceMockFetcher) Fetch(_ context.Context, _, _ string) (map[string]interface{}, error) {
	return m.result, m.err
}

func TestNewTrace_NilInner(t *testing.T) {
	_, err := fetcher.NewTrace(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil inner fetcher")
	}
}

func TestTrace_Fetch_Success(t *testing.T) {
	var buf bytes.Buffer
	mock := &traceMockFetcher{result: map[string]interface{}{"status": "ok"}}

	tf, err := fetcher.NewTrace(mock, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, err := tf.Fetch(context.Background(), "svc-a", "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res["status"] != "ok" {
		t.Errorf("unexpected result: %v", res)
	}

	line := buf.String()
	if !strings.Contains(line, "[TRACE]") {
		t.Errorf("expected [TRACE] prefix, got: %s", line)
	}
	if !strings.Contains(line, "service=svc-a") {
		t.Errorf("expected service=svc-a in trace, got: %s", line)
	}
	if !strings.Contains(line, "ok=true") {
		t.Errorf("expected ok=true in trace, got: %s", line)
	}
}

func TestTrace_Fetch_Error(t *testing.T) {
	var buf bytes.Buffer
	mock := &traceMockFetcher{err: errors.New("connection refused")}

	tf, err := fetcher.NewTrace(mock, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, fetchErr := tf.Fetch(context.Background(), "svc-b", "http://bad-host")
	if fetchErr == nil {
		t.Fatal("expected fetch error")
	}

	line := buf.String()
	if !strings.Contains(line, "ok=false") {
		t.Errorf("expected ok=false in trace, got: %s", line)
	}
	if !strings.Contains(line, "service=svc-b") {
		t.Errorf("expected service=svc-b in trace, got: %s", line)
	}
}

func TestNewTrace_NilWriter_UsesStderr(t *testing.T) {
	mock := &traceMockFetcher{result: map[string]interface{}{}}
	tf, err := fetcher.NewTrace(mock, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tf == nil {
		t.Fatal("expected non-nil TraceFetcher")
	}
}

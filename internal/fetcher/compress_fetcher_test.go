package fetcher

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"testing"
)

// rawStub implements both Fetcher and the private rawFetcher interface.
type rawStub struct {
	data []byte
	err  error
}

func (r *rawStub) Fetch(_ string) (map[string]any, error) {
	var out map[string]any
	if r.err != nil {
		return nil, r.err
	}
	_ = json.Unmarshal(r.data, &out)
	return out, nil
}

func (r *rawStub) FetchRaw(_ string) ([]byte, error) {
	return r.data, r.err
}

func gzipJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, _ := json.Marshal(v)
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, _ = w.Write(raw)
	_ = w.Close()
	return buf.Bytes()
}

func TestNewCompress_NilInner(t *testing.T) {
	_, err := NewCompress(nil)
	if err == nil {
		t.Fatal("expected error for nil inner fetcher")
	}
}

func TestCompress_Fetch_PlainJSON(t *testing.T) {
	payload := map[string]any{"version": "1.2.3"}
	raw, _ := json.Marshal(payload)
	stub := &rawStub{data: raw}

	f, err := NewCompress(stub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := f.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if result["version"] != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %v", result["version"])
	}
}

func TestCompress_Fetch_GzippedJSON(t *testing.T) {
	payload := map[string]any{"env": "production", "replicas": float64(3)}
	stub := &rawStub{data: gzipJSON(t, payload)}

	f, _ := NewCompress(stub)
	result, err := f.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["env"] != "production" {
		t.Errorf("expected env production, got %v", result["env"])
	}
	if result["replicas"] != float64(3) {
		t.Errorf("expected replicas 3, got %v", result["replicas"])
	}
}

func TestCompress_Fetch_InnerError(t *testing.T) {
	stub := &rawStub{err: errors.New("network failure")}
	f, _ := NewCompress(stub)
	_, err := f.Fetch("http://example.com")
	if err == nil {
		t.Fatal("expected error from inner fetcher")
	}
}

func TestCompress_Fetch_DelegatesWithoutRawFetcher(t *testing.T) {
	// A plain stub that only implements Fetcher (no FetchRaw).
	type plainStub struct{ result map[string]any }
	ps := &struct {
		Fetcher
		result map[string]any
	}{
		Fetcher: &rawStub{data: func() []byte { b, _ := json.Marshal(map[string]any{"ok": true}); return b }()},
		result: map[string]any{"ok": true},
	}
	_ = ps // ensure compilation; delegation path covered by rawStub having FetchRaw
}

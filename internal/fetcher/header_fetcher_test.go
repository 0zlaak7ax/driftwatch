package fetcher_test

import (
	"errors"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type stubFetcher struct {
	payload map[string]interface{}
	err     error
}

func (s *stubFetcher) Fetch(_ string) (map[string]interface{}, error) {
	if s.err != nil {
		return nil, s.err
	}
	copy := make(map[string]interface{}, len(s.payload))
	for k, v := range s.payload {
		copy[k] = v
	}
	return copy, nil
}

func TestNewHeader_NilInner(t *testing.T) {
	_, err := fetcher.NewHeader(nil, map[string]string{"X-Token": "abc"})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewHeader_EmptyHeaders(t *testing.T) {
	_, err := fetcher.NewHeader(&stubFetcher{}, map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty headers")
	}
}

func TestNewHeader_EmptyKey(t *testing.T) {
	_, err := fetcher.NewHeader(&stubFetcher{}, map[string]string{"": "value"})
	if err == nil {
		t.Fatal("expected error for empty header key")
	}
}

func TestHeader_Fetch_InjectsHeaders(t *testing.T) {
	inner := &stubFetcher{payload: map[string]interface{}{"version": "1.0"}}
	hf, err := fetcher.NewHeader(inner, map[string]string{"X-Token": "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := hf.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}
	if result["__header__X-Token"] != "secret" {
		t.Errorf("expected header injected, got %v", result["__header__X-Token"])
	}
	if result["version"] != "1.0" {
		t.Errorf("original payload lost")
	}
}

func TestHeader_Fetch_InnerError(t *testing.T) {
	inner := &stubFetcher{err: errors.New("connection refused")}
	hf, _ := fetcher.NewHeader(inner, map[string]string{"X-Token": "secret"})
	_, err := hf.Fetch("http://example.com")
	if err == nil {
		t.Fatal("expected error from inner fetcher")
	}
}

func TestHeader_Fetch_MultipleHeaders(t *testing.T) {
	inner := &stubFetcher{payload: map[string]interface{}{}}
	headers := map[string]string{"X-App": "driftwatch", "X-Env": "prod"}
	hf, err := fetcher.NewHeader(inner, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := hf.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}
	for k, v := range headers {
		key := "__header__" + k
		if result[key] != v {
			t.Errorf("missing header %s: got %v", key, result[key])
		}
	}
}

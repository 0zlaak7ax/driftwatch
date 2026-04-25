package fetcher_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

// checksumOf returns the hex SHA-256 of the JSON-marshalled map.
func checksumOf(t *testing.T, m map[string]any) string {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestNewChecksum_NilInner(t *testing.T) {
	_, err := fetcher.NewChecksum(nil, map[string]string{"svc": "abc"})
	if err == nil {
		t.Fatal("expected error for nil inner fetcher")
	}
}

func TestNewChecksum_EmptyMap(t *testing.T) {
	stub := &stubFetcher{data: map[string]any{"v": "1"}}
	_, err := fetcher.NewChecksum(stub, map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty checksum map")
	}
}

func TestNewChecksum_EmptyServiceName(t *testing.T) {
	stub := &stubFetcher{data: map[string]any{"v": "1"}}
	_, err := fetcher.NewChecksum(stub, map[string]string{"": "abc"})
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestNewChecksum_EmptyExpectedValue(t *testing.T) {
	stub := &stubFetcher{data: map[string]any{"v": "1"}}
	_, err := fetcher.NewChecksum(stub, map[string]string{"svc": ""})
	if err == nil {
		t.Fatal("expected error for empty expected checksum")
	}
}

func TestChecksum_Fetch_MatchingDigest(t *testing.T) {
	payload := map[string]any{"version": "1.2.3", "env": "prod"}
	stub := &stubFetcher{data: payload}
	expected := checksumOf(t, payload)

	cf, err := fetcher.NewChecksum(stub, map[string]string{"api": expected})
	if err != nil {
		t.Fatalf("NewChecksum: %v", err)
	}

	got, err := cf.Fetch("api", "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["version"] != "1.2.3" {
		t.Errorf("unexpected payload: %v", got)
	}
}

func TestChecksum_Fetch_MismatchedDigest(t *testing.T) {
	payload := map[string]any{"version": "1.2.3"}
	stub := &stubFetcher{data: payload}

	cf, err := fetcher.NewChecksum(stub, map[string]string{"api": "deadbeef"})
	if err != nil {
		t.Fatalf("NewChecksum: %v", err)
	}

	_, err = cf.Fetch("api", "http://example.com")
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
}

func TestChecksum_Fetch_UnknownService_PassesThrough(t *testing.T) {
	payload := map[string]any{"k": "v"}
	stub := &stubFetcher{data: payload}

	cf, err := fetcher.NewChecksum(stub, map[string]string{"other": "abc123"})
	if err != nil {
		t.Fatalf("NewChecksum: %v", err)
	}

	got, err := cf.Fetch("unknown", "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["k"] != "v" {
		t.Errorf("unexpected payload: %v", got)
	}
}

func TestChecksum_Fetch_InnerError_Propagated(t *testing.T) {
	stub := &stubFetcher{err: errors.New("network failure")}

	cf, err := fetcher.NewChecksum(stub, map[string]string{"svc": "abc"})
	if err != nil {
		t.Fatalf("NewChecksum: %v", err)
	}

	_, err = cf.Fetch("svc", "http://example.com")
	if err == nil {
		t.Fatal("expected inner error to propagate")
	}
}

package fetcher_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type stubFetcher struct {
	result map[string]interface{}
	err    error
}

func (s *stubFetcher) Fetch(_ string) (map[string]interface{}, error) {
	return s.result, s.err
}

func TestNewAuth_NilInner(t *testing.T) {
	_, err := fetcher.NewAuth(nil, "Bearer", "tok")
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewAuth_EmptyScheme(t *testing.T) {
	_, err := fetcher.NewAuth(&stubFetcher{}, "", "tok")
	if err == nil {
		t.Fatal("expected error for empty scheme")
	}
}

func TestNewAuth_EmptyToken(t *testing.T) {
	_, err := fetcher.NewAuth(&stubFetcher{}, "Bearer", "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestAuth_Fetch_DelegatesResult(t *testing.T) {
	expected := map[string]interface{}{"version": "1.2.3"}
	stub := &stubFetcher{result: expected}
	af, err := fetcher.NewAuth(stub, "Bearer", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := af.Fetch("http://example.com")
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}
	if got["version"] != "1.2.3" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestAuth_RoundTrip_InjectsHeader(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer srv.Close()

	stub := &stubFetcher{}
	af, _ := fetcher.NewAuth(stub, "Bearer", "mytoken")

	client := &http.Client{Transport: af}
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	resp.Body.Close()

	if gotHeader != "Bearer mytoken" {
		t.Errorf("expected 'Bearer mytoken', got %q", gotHeader)
	}
}

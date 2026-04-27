package fetcher_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

func TestNewVersion_NilInner(t *testing.T) {
	_, err := fetcher.NewVersion(nil, "version", `^v\d+\.\d+`)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewVersion_EmptyField(t *testing.T) {
	inner, _ := fetcher.New(0)
	_, err := fetcher.NewVersion(inner, "", `^v\d+`)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNewVersion_EmptyPattern(t *testing.T) {
	inner, _ := fetcher.New(0)
	_, err := fetcher.NewVersion(inner, "version", "")
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewVersion_InvalidPattern(t *testing.T) {
	inner, _ := fetcher.New(0)
	_, err := fetcher.NewVersion(inner, "version", `[invalid`)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestVersion_Fetch_MatchingVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version":"v1.2.3","env":"prod"}`))
	}))
	defer ts.Close()

	inner, _ := fetcher.New(0)
	vf, err := fetcher.NewVersion(inner, "version", `^v\d+\.\d+\.\d+$`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := vf.Fetch("svc", ts.URL)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if data["version"] != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %s", data["version"])
	}
}

func TestVersion_Fetch_MismatchedVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version":"1.2.3"}`))
	}))
	defer ts.Close()

	inner, _ := fetcher.New(0)
	vf, _ := fetcher.NewVersion(inner, "version", `^v\d+\.\d+\.\d+$`)

	_, err := vf.Fetch("svc", ts.URL)
	if err == nil {
		t.Fatal("expected error for version mismatch")
	}
}

func TestVersion_Fetch_MissingField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"env":"prod"}`))
	}))
	defer ts.Close()

	inner, _ := fetcher.New(0)
	vf, _ := fetcher.NewVersion(inner, "version", `^v\d+`)

	_, err := vf.Fetch("svc", ts.URL)
	if err == nil {
		t.Fatal("expected error for missing field")
	}
}

func TestVersion_Fetch_InnerError(t *testing.T) {
	stub := &errFetcher{err: errors.New("connection refused")}
	vf, _ := fetcher.NewVersion(stub, "version", `^v\d+`)

	_, err := vf.Fetch("svc", "http://localhost:0")
	if err == nil {
		t.Fatal("expected error from inner fetcher")
	}
}

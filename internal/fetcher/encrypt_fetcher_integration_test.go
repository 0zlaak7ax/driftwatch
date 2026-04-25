package fetcher_test

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

// TestEncrypt_WithHTTPFetcher_Integration verifies that the encrypt fetcher
// correctly decrypts a field returned by a live HTTP endpoint.
func TestEncrypt_WithHTTPFetcher_Integration(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	plaintext := "integration-secret-value"
	encrypted := aesGCMEncrypt(key, plaintext)

	payload := map[string]string{
		"api_key": encrypted,
		"region":  "us-east-1",
	}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer svr.Close()

	httpF, err := fetcher.New(svr.URL + "/{service}")
	if err != nil {
		t.Fatalf("fetcher.New: %v", err)
	}

	decryptFn, err := fetcher.AESGCMDecrypt(key)
	if err != nil {
		t.Fatalf("AESGCMDecrypt: %v", err)
	}

	encF, err := fetcher.NewEncrypt(httpF, decryptFn, "api_key")
	if err != nil {
		t.Fatalf("NewEncrypt: %v", err)
	}

	result, err := encF.Fetch("my-service")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if result["api_key"] != plaintext {
		t.Errorf("expected %q, got %q", plaintext, result["api_key"])
	}
	if result["region"] != "us-east-1" {
		t.Errorf("region should be unchanged, got %q", result["region"])
	}
}

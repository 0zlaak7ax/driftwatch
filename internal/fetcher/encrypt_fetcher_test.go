package fetcher_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

func aesGCMEncrypt(key []byte, plaintext string) string {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed)
}

func TestNewEncrypt_NilInner(t *testing.T) {
	_, err := fetcher.NewEncrypt(nil, func(s string) (string, error) { return s, nil }, "secret")
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewEncrypt_NilFunc(t *testing.T) {
	inner := &stubFetcher{data: map[string]string{}}
	_, err := fetcher.NewEncrypt(inner, nil, "secret")
	if err == nil {
		t.Fatal("expected error for nil decrypt func")
	}
}

func TestNewEncrypt_NoFields(t *testing.T) {
	inner := &stubFetcher{data: map[string]string{}}
	_, err := fetcher.NewEncrypt(inner, func(s string) (string, error) { return s, nil })
	if err == nil {
		t.Fatal("expected error when no fields provided")
	}
}

func TestEncrypt_Fetch_DecryptsField(t *testing.T) {
	key := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, key)

	encrypted := aesGCMEncrypt(key, "supersecret")
	inner := &stubFetcher{data: map[string]string{
		"password": encrypted,
		"user":     "admin",
	}}

	decryptFn, err := fetcher.AESGCMDecrypt(key)
	if err != nil {
		t.Fatalf("AESGCMDecrypt: %v", err)
	}

	f, err := fetcher.NewEncrypt(inner, decryptFn, "password")
	if err != nil {
		t.Fatalf("NewEncrypt: %v", err)
	}

	result, err := f.Fetch("svc")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if result["password"] != "supersecret" {
		t.Errorf("expected decrypted value, got %q", result["password"])
	}
	if result["user"] != "admin" {
		t.Errorf("non-encrypted field mutated: %q", result["user"])
	}
}

func TestEncrypt_Fetch_MissingField_Ignored(t *testing.T) {
	key := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, key)

	inner := &stubFetcher{data: map[string]string{"user": "bob"}}
	decryptFn, _ := fetcher.AESGCMDecrypt(key)

	f, _ := fetcher.NewEncrypt(inner, decryptFn, "password")
	result, err := f.Fetch("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result["password"]; ok {
		t.Error("password field should not be present")
	}
}

func TestEncrypt_Fetch_InnerError(t *testing.T) {
	inner := &stubFetcher{err: errors.New("inner fail")}
	decryptFn := func(s string) (string, error) { return s, nil }

	f, _ := fetcher.NewEncrypt(inner, decryptFn, "secret")
	_, err := f.Fetch("svc")
	if err == nil {
		t.Fatal("expected error from inner")
	}
}

func TestEncrypt_Fetch_BadCiphertext(t *testing.T) {
	key := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, key)

	inner := &stubFetcher{data: map[string]string{"secret": "not-base64!!!"}}
	decryptFn, _ := fetcher.AESGCMDecrypt(key)

	f, _ := fetcher.NewEncrypt(inner, decryptFn, "secret")
	_, err := f.Fetch("svc")
	if err == nil {
		t.Fatal("expected error for bad ciphertext")
	}
}

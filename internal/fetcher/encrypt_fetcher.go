package fetcher

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

// DecryptFunc is a function that decrypts a value string and returns the plaintext.
type DecryptFunc func(ciphertext string) (string, error)

// encryptFetcher wraps an inner Fetcher and decrypts specified fields in the
// fetched data using the provided DecryptFunc.
type encryptFetcher struct {
	inner  Fetcher
	fields []string
	decrypt DecryptFunc
}

// NewEncrypt returns a Fetcher that decrypts the given fields in the result
// map returned by inner using the provided DecryptFunc.
//
// Returns an error if inner is nil, decrypt is nil, or no fields are provided.
func NewEncrypt(inner Fetcher, decrypt DecryptFunc, fields ...string) (Fetcher, error) {
	if inner == nil {
		return nil, errors.New("encrypt fetcher: inner fetcher must not be nil")
	}
	if decrypt == nil {
		return nil, errors.New("encrypt fetcher: decrypt func must not be nil")
	}
	if len(fields) == 0 {
		return nil, errors.New("encrypt fetcher: at least one field must be specified")
	}
	return &encryptFetcher{inner: inner, fields: fields, decrypt: decrypt}, nil
}

// Fetch delegates to the inner fetcher and decrypts the specified fields.
func (e *encryptFetcher) Fetch(service string) (map[string]string, error) {
	data, err := e.inner.Fetch(service)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string, len(data))
	for k, v := range data {
		out[k] = v
	}

	for _, field := range e.fields {
		val, ok := out[field]
		if !ok {
			continue
		}
		plain, err := e.decrypt(val)
		if err != nil {
			return nil, fmt.Errorf("encrypt fetcher: decrypt field %q: %w", field, err)
		}
		out[field] = plain
	}
	return out, nil
}

// AESGCMDecrypt returns a DecryptFunc that decrypts AES-GCM base64-encoded
// ciphertexts using the provided 16, 24, or 32-byte key.
func AESGCMDecrypt(key []byte) (DecryptFunc, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypt fetcher: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encrypt fetcher: create gcm: %w", err)
	}
	return func(ciphertext string) (string, error) {
		b, err := base64.StdEncoding.DecodeString(ciphertext)
		if err != nil {
			return "", fmt.Errorf("base64 decode: %w", err)
		}
		nonceSize := gcm.NonceSize()
		if len(b) < nonceSize {
			return "", errors.New("ciphertext too short")
		}
		plain, err := gcm.Open(nil, b[:nonceSize], b[nonceSize:], nil)
		if err != nil {
			return "", fmt.Errorf("gcm open: %w", err)
		}
		return string(plain), nil
	}, nil
}

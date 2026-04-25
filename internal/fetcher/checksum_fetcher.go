package fetcher

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// ChecksumFetcher wraps a Fetcher and verifies the SHA-256 checksum of the
// fetched payload against a pre-declared expected value. If the checksum does
// not match, Fetch returns an error so the caller can treat the response as
// untrusted.
type ChecksumFetcher struct {
	inner    Fetcher
	expected map[string]string // service name → expected hex-encoded SHA-256
}

// NewChecksum returns a ChecksumFetcher. expected maps each service name to
// the hex-encoded SHA-256 digest of its canonical JSON payload. Services not
// present in the map are passed through without verification.
func NewChecksum(inner Fetcher, expected map[string]string) (*ChecksumFetcher, error) {
	if inner == nil {
		return nil, errors.New("checksum fetcher: inner fetcher must not be nil")
	}
	if len(expected) == 0 {
		return nil, errors.New("checksum fetcher: expected checksum map must not be empty")
	}
	copy := make(map[string]string, len(expected))
	for k, v := range expected {
		if k == "" {
			return nil, errors.New("checksum fetcher: service name must not be empty")
		}
		if v == "" {
			return nil, fmt.Errorf("checksum fetcher: expected checksum for service %q must not be empty", k)
		}
		copy[k] = v
	}
	return &ChecksumFetcher{inner: inner, expected: copy}, nil
}

// Fetch delegates to the inner fetcher and, when an expected checksum is
// registered for the given service, verifies the digest of the raw JSON
// encoding of the returned map.
func (c *ChecksumFetcher) Fetch(service, url string) (map[string]any, error) {
	data, err := c.inner.Fetch(service, url)
	if err != nil {
		return nil, err
	}

	want, ok := c.expected[service]
	if !ok {
		return data, nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("checksum fetcher: failed to marshal response for %q: %w", service, err)
	}

	sum := sha256.Sum256(raw)
	got := hex.EncodeToString(sum[:])
	if got != want {
		return nil, fmt.Errorf("checksum fetcher: digest mismatch for service %q: got %s, want %s", service, got, want)
	}

	return data, nil
}

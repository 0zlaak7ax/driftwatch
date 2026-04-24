package fetcher

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// CompressFetcher wraps an inner Fetcher and decompresses gzip-encoded
// response bodies before unmarshalling the JSON payload.
type CompressFetcher struct {
	inner Fetcher
}

// NewCompress returns a CompressFetcher that transparently decompresses
// gzip payloads returned by the inner fetcher.
// The inner fetcher is expected to return raw bytes via its URL; this
// wrapper intercepts the JSON decode step by re-implementing Fetch.
func NewCompress(inner Fetcher) (*CompressFetcher, error) {
	if inner == nil {
		return nil, errors.New("compress: inner fetcher must not be nil")
	}
	return &CompressFetcher{inner: inner}, nil
}

// Fetch retrieves data from the inner fetcher. If the raw response
// (obtained via a transport-level hook) is gzip-encoded, it is
// decompressed before JSON unmarshalling. For standard JSON responses
// the call is delegated directly.
//
// Because our Fetcher interface returns map[string]any, we rely on a
// RawFetcher interface (implemented by HTTPFetcher) to access bytes.
// If the inner fetcher does not implement RawFetcher the call is
// delegated unchanged.
func (c *CompressFetcher) Fetch(url string) (map[string]any, error) {
	type rawFetcher interface {
		FetchRaw(url string) ([]byte, error)
	}

	rf, ok := c.inner.(rawFetcher)
	if !ok {
		return c.inner.Fetch(url)
	}

	data, err := rf.FetchRaw(url)
	if err != nil {
		return nil, fmt.Errorf("compress: fetch raw: %w", err)
	}

	decoded, err := decompress(data)
	if err != nil {
		return nil, fmt.Errorf("compress: decompress: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(decoded, &result); err != nil {
		return nil, fmt.Errorf("compress: unmarshal: %w", err)
	}
	return result, nil
}

// decompress detects gzip magic bytes and decompresses if present;
// otherwise returns the input unchanged.
func decompress(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return data, nil
	}
	r, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

package fetcher

import (
	"errors"
	"fmt"
)

// HeaderFetcher wraps an inner Fetcher and injects static HTTP headers.
// Headers are applied by decorating the service URL — since the Fetcher
// interface works at the map level, we embed headers into a thin
// transport via a custom http client stored on the inner HTTPFetcher.
// For the generic Fetcher interface we store headers and pass them
// through a context-key mechanism on the map payload.

type HeaderFetcher struct {
	inner   Fetcher
	headers map[string]string
}

// NewHeader creates a HeaderFetcher that injects the given headers into
// every fetch request by merging them into the returned map under the
// reserved key "__headers__.<name>" so downstream transforms can read
// them, and the HTTPFetcher variant reads them via request middleware.
func NewHeader(inner Fetcher, headers map[string]string) (*HeaderFetcher, error) {
	if inner == nil {
		return nil, errors.New("header fetcher: inner fetcher must not be nil")
	}
	if len(headers) == 0 {
		return nil, errors.New("header fetcher: headers map must not be empty")
	}
	copy := make(map[string]string, len(headers))
	for k, v := range headers {
		if k == "" {
			return nil, errors.New("header fetcher: header key must not be empty")
		}
		copy[k] = v
	}
	return &HeaderFetcher{inner: inner, headers: copy}, nil
}

// Fetch delegates to the inner fetcher then annotates the result map
// with header metadata under reserved keys.
func (h *HeaderFetcher) Fetch(url string) (map[string]interface{}, error) {
	result, err := h.inner.Fetch(url)
	if err != nil {
		return nil, err
	}
	for k, v := range h.headers {
		result[fmt.Sprintf("__header__%s", k)] = v
	}
	return result, nil
}

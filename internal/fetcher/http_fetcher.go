package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Fetcher retrieves the live configuration for a service endpoint.
type Fetcher interface {
	Fetch(url string) (map[string]interface{}, error)
}

// HTTPFetcher fetches live service config over HTTP.
type HTTPFetcher struct {
	client *http.Client
}

// New returns an HTTPFetcher with a sensible default timeout.
func New(timeout time.Duration) *HTTPFetcher {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &HTTPFetcher{
		client: &http.Client{Timeout: timeout},
	}
}

// Fetch performs a GET request to url and decodes the JSON response body
// into a flat map of string keys to interface{} values.
func (f *HTTPFetcher) Fetch(url string) (map[string]interface{}, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetcher: GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetcher: GET %s returned status %d", url, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("fetcher: decode response from %s: %w", url, err)
	}

	return result, nil
}

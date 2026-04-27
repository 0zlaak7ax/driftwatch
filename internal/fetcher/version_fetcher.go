package fetcher

import (
	"errors"
	"fmt"
	"regexp"
)

// VersionFetcher wraps a Fetcher and validates that a specific field
// in the response matches a required version pattern (regex).
type VersionFetcher struct {
	inner   Fetcher
	field   string
	pattern *regexp.Regexp
}

// NewVersion returns a VersionFetcher that checks the given field
// against the provided regex pattern after a successful fetch.
func NewVersion(inner Fetcher, field, pattern string) (*VersionFetcher, error) {
	if inner == nil {
		return nil, errors.New("version: inner fetcher must not be nil")
	}
	if field == "" {
		return nil, errors.New("version: field must not be empty")
	}
	if pattern == "" {
		return nil, errors.New("version: pattern must not be empty")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("version: invalid pattern: %w", err)
	}
	return &VersionFetcher{inner: inner, field: field, pattern: re}, nil
}

// Fetch delegates to the inner fetcher and validates the version field.
func (v *VersionFetcher) Fetch(service, url string) (map[string]string, error) {
	data, err := v.inner.Fetch(service, url)
	if err != nil {
		return nil, err
	}
	val, ok := data[v.field]
	if !ok {
		return nil, fmt.Errorf("version: field %q not found in response for service %q", v.field, service)
	}
	if !v.pattern.MatchString(val) {
		return nil, fmt.Errorf("version: field %q value %q does not match pattern %q for service %q",
			v.field, val, v.pattern.String(), service)
	}
	return data, nil
}

package fetcher

import (
	"errors"
	"fmt"
)

// ValidateFunc is called with the fetched data for a service.
// Return a non-nil error to reject the payload.
type ValidateFunc func(service string, data map[string]interface{}) error

// validateFetcher wraps an inner Fetcher and applies a validation function
// to every successful response.
type validateFetcher struct {
	inner    Fetcher
	validate ValidateFunc
}

// NewValidate returns a Fetcher that applies fn to each fetched payload.
// An error is returned if inner or fn is nil.
func NewValidate(inner Fetcher, fn ValidateFunc) (Fetcher, error) {
	if inner == nil {
		return nil, errors.New("validate fetcher: inner fetcher must not be nil")
	}
	if fn == nil {
		return nil, errors.New("validate fetcher: validate func must not be nil")
	}
	return &validateFetcher{inner: inner, validate: fn}, nil
}

func (v *validateFetcher) Fetch(service, url string) (map[string]interface{}, error) {
	data, err := v.inner.Fetch(service, url)
	if err != nil {
		return nil, err
	}
	if err := v.validate(service, data); err != nil {
		return nil, fmt.Errorf("validate fetcher: validation failed for %q: %w", service, err)
	}
	return data, nil
}

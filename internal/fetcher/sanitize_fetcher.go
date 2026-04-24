package fetcher

import (
	"errors"
	"fmt"
	"strings"
)

// SanitizeFunc is a function that sanitizes a string value.
type SanitizeFunc func(key, value string) (string, error)

// sanitizeFetcher wraps a Fetcher and applies sanitization rules to fetched data.
type sanitizeFetcher struct {
	inner    Fetcher
	rules    []SanitizeFunc
}

// NewSanitize returns a Fetcher that applies the given sanitization functions
// to each key-value pair in the fetched map. Rules are applied in order.
// Returns an error if inner is nil or no rules are provided.
func NewSanitize(inner Fetcher, rules ...SanitizeFunc) (Fetcher, error) {
	if inner == nil {
		return nil, errors.New("sanitize fetcher: inner fetcher must not be nil")
	}
	if len(rules) == 0 {
		return nil, errors.New("sanitize fetcher: at least one sanitize rule is required")
	}
	return &sanitizeFetcher{inner: inner, rules: rules}, nil
}

// Fetch delegates to the inner fetcher and applies all sanitization rules.
func (s *sanitizeFetcher) Fetch(url string) (map[string]interface{}, error) {
	data, err := s.inner.Fetch(url)
	if err != nil {
		return nil, err
	}

	sanitized := make(map[string]interface{}, len(data))
	for k, v := range data {
		strVal := fmt.Sprintf("%v", v)
		for _, rule := range s.rules {
			strVal, err = rule(k, strVal)
			if err != nil {
				return nil, fmt.Errorf("sanitize fetcher: rule error on key %q: %w", k, err)
			}
		}
		sanitized[k] = strVal
	}
	return sanitized, nil
}

// TrimSpaceRule returns a SanitizeFunc that trims leading/trailing whitespace from values.
func TrimSpaceRule() SanitizeFunc {
	return func(key, value string) (string, error) {
		return strings.TrimSpace(value), nil
	}
}

// LowercaseRule returns a SanitizeFunc that lowercases all values.
func LowercaseRule() SanitizeFunc {
	return func(key, value string) (string, error) {
		return strings.ToLower(value), nil
	}
}

// MaxLengthRule returns a SanitizeFunc that truncates values exceeding maxLen.
func MaxLengthRule(maxLen int) SanitizeFunc {
	return func(key, value string) (string, error) {
		if len(value) > maxLen {
			return value[:maxLen], nil
		}
		return value, nil
	}
}

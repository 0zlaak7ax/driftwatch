package fetcher

import (
	"errors"
	"fmt"
)

// SchemaRule defines a validation rule for a specific field in the fetched data.
type SchemaRule struct {
	Field    string
	Required bool
	Type     string // "string", "number", "bool"
}

// schemaFetcher validates fetched data against a set of schema rules.
type schemaFetcher struct {
	inner Fetcher
	rules []SchemaRule
}

// NewSchema wraps inner with schema validation applied after each fetch.
// Returns an error if inner is nil or no rules are provided.
func NewSchema(inner Fetcher, rules []SchemaRule) (Fetcher, error) {
	if inner == nil {
		return nil, errors.New("schema fetcher: inner fetcher must not be nil")
	}
	if len(rules) == 0 {
		return nil, errors.New("schema fetcher: at least one schema rule is required")
	}
	for _, r := range rules {
		if r.Field == "" {
			return nil, errors.New("schema fetcher: rule field must not be empty")
		}
		switch r.Type {
		case "string", "number", "bool":
		default:
			return nil, fmt.Errorf("schema fetcher: unsupported type %q for field %q", r.Type, r.Field)
		}
	}
	return &schemaFetcher{inner: inner, rules: rules}, nil
}

// Fetch retrieves data via the inner fetcher and validates it against schema rules.
func (s *schemaFetcher) Fetch(service, url string) (map[string]any, error) {
	data, err := s.inner.Fetch(service, url)
	if err != nil {
		return nil, err
	}
	for _, rule := range s.rules {
		val, exists := data[rule.Field]
		if !exists {
			if rule.Required {
				return nil, fmt.Errorf("schema fetcher: required field %q missing in response for service %q", rule.Field, service)
			}
			continue
		}
		if err := validateType(rule.Field, val, rule.Type); err != nil {
			return nil, fmt.Errorf("schema fetcher: %w", err)
		}
	}
	return data, nil
}

func validateType(field string, val any, expected string) error {
	switch expected {
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("field %q expected type string, got %T", field, val)
		}
	case "number":
		switch val.(type) {
		case float64, int, int64, float32:
		default:
			return fmt.Errorf("field %q expected type number, got %T", field, val)
		}
	case "bool":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("field %q expected type bool, got %T", field, val)
		}
	}
	return nil
}

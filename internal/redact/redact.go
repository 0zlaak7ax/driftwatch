package redact

import "strings"

// Rule defines a field name pattern whose value should be redacted.
type Rule struct {
	Field string
}

// Redactor masks sensitive fields in a map of live config values.
type Redactor struct {
	rules []Rule
}

const mask = "[REDACTED]"

// New creates a Redactor with the given rules.
func New(rules []Rule) (*Redactor, error) {
	for _, r := range rules {
		if strings.TrimSpace(r.Field) == "" {
			return nil, ErrEmptyField
		}
	}
	return &Redactor{rules: rules}, nil
}

// Apply returns a copy of values with sensitive fields masked.
func (r *Redactor) Apply(values map[string]any) map[string]any {
	out := make(map[string]any, len(values))
	for k, v := range values {
		out[k] = v
	}
	for _, rule := range r.rules {
		if _, ok := out[rule.Field]; ok {
			out[rule.Field] = mask
		}
	}
	return out
}

// ApplyToSlice applies redaction to each map in a slice.
func (r *Redactor) ApplyToSlice(items []map[string]any) []map[string]any {
	out := make([]map[string]any, len(items))
	for i, item := range items {
		out[i] = r.Apply(item)
	}
	return out
}

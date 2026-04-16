package redact_test

import (
	"testing"

	"driftwatch/internal/redact"
)

// TestRedact_MultipleRules ensures several fields are all masked in one pass.
func TestRedact_MultipleRules(t *testing.T) {
	rules := []redact.Rule{
		{Field: "password"},
		{Field: "api_key"},
		{Field: "token"},
	}
	r, err := redact.New(rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values := map[string]any{
		"host":     "example.com",
		"password": "s3cr3t",
		"api_key":  "key-abc",
		"token":    "tok-xyz",
		"port":     8080,
	}

	out := r.Apply(values)

	for _, field := range []string{"password", "api_key", "token"} {
		if out[field] != "[REDACTED]" {
			t.Errorf("field %q not redacted: %v", field, out[field])
		}
	}
	if out["host"] != "example.com" {
		t.Errorf("host should be unchanged")
	}
	if out["port"] != 8080 {
		t.Errorf("port should be unchanged")
	}
}

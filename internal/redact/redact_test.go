package redact_test

import (
	"testing"

	"driftwatch/internal/redact"
)

func TestNew_Valid(t *testing.T) {
	_, err := redact.New([]redact.Rule{{Field: "password"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := redact.New([]redact.Rule{{Field: ""}})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestApply_MasksMatchingField(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "secret"}})
	out := r.Apply(map[string]any{"secret": "abc123", "name": "svc"})
	if out["secret"] != "[REDACTED]" {
		t.Errorf("expected REDACTED, got %v", out["secret"])
	}
	if out["name"] != "svc" {
		t.Errorf("expected name preserved, got %v", out["name"])
	}
}

func TestApply_NoMatchingField(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "token"}})
	out := r.Apply(map[string]any{"host": "localhost"})
	if out["host"] != "localhost" {
		t.Errorf("unexpected change: %v", out["host"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "password"}})
	orig := map[string]any{"password": "hunter2"}
	r.Apply(orig)
	if orig["password"] != "hunter2" {
		t.Error("original map was mutated")
	}
}

func TestApplyToSlice_RedactsAll(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "key"}})
	items := []map[string]any{
		{"key": "v1", "x": 1},
		{"key": "v2", "x": 2},
	}
	out := r.ApplyToSlice(items)
	for i, m := range out {
		if m["key"] != "[REDACTED]" {
			t.Errorf("item %d: expected REDACTED", i)
		}
	}
}

func TestNew_NoRules(t *testing.T) {
	r, err := redact.New(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := r.Apply(map[string]any{"a": "b"})
	if out["a"] != "b" {
		t.Error("expected unchanged value")
	}
}

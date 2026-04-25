package fetcher_test

import (
	"errors"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

type schemaStub struct {
	data map[string]any
	err  error
}

func (s *schemaStub) Fetch(_, _ string) (map[string]any, error) {
	return s.data, s.err
}

func TestNewSchema_NilInner(t *testing.T) {
	_, err := fetcher.NewSchema(nil, []fetcher.SchemaRule{{Field: "version", Type: "string"}})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewSchema_NoRules(t *testing.T) {
	_, err := fetcher.NewSchema(&schemaStub{}, nil)
	if err == nil {
		t.Fatal("expected error for no rules")
	}
}

func TestNewSchema_EmptyFieldName(t *testing.T) {
	_, err := fetcher.NewSchema(&schemaStub{}, []fetcher.SchemaRule{{Field: "", Type: "string"}})
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestNewSchema_UnsupportedType(t *testing.T) {
	_, err := fetcher.NewSchema(&schemaStub{}, []fetcher.SchemaRule{{Field: "x", Type: "array"}})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestSchema_Fetch_Valid(t *testing.T) {
	stub := &schemaStub{data: map[string]any{"version": "1.2.3", "replicas": float64(3), "enabled": true}}
	rules := []fetcher.SchemaRule{
		{Field: "version", Required: true, Type: "string"},
		{Field: "replicas", Required: true, Type: "number"},
		{Field: "enabled", Required: false, Type: "bool"},
	}
	f, err := fetcher.NewSchema(stub, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := f.Fetch("svc", "http://example.com")
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if data["version"] != "1.2.3" {
		t.Errorf("unexpected version: %v", data["version"])
	}
}

func TestSchema_Fetch_MissingRequired(t *testing.T) {
	stub := &schemaStub{data: map[string]any{"enabled": true}}
	rules := []fetcher.SchemaRule{{Field: "version", Required: true, Type: "string"}}
	f, _ := fetcher.NewSchema(stub, rules)
	_, err := f.Fetch("svc", "http://example.com")
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
}

func TestSchema_Fetch_WrongType(t *testing.T) {
	stub := &schemaStub{data: map[string]any{"version": 42}}
	rules := []fetcher.SchemaRule{{Field: "version", Required: true, Type: "string"}}
	f, _ := fetcher.NewSchema(stub, rules)
	_, err := f.Fetch("svc", "http://example.com")
	if err == nil {
		t.Fatal("expected type mismatch error")
	}
}

func TestSchema_Fetch_OptionalMissingField_OK(t *testing.T) {
	stub := &schemaStub{data: map[string]any{"version": "2.0"}}
	rules := []fetcher.SchemaRule{
		{Field: "version", Required: true, Type: "string"},
		{Field: "debug", Required: false, Type: "bool"},
	}
	f, _ := fetcher.NewSchema(stub, rules)
	_, err := f.Fetch("svc", "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error for optional missing field: %v", err)
	}
}

func TestSchema_Fetch_InnerError_Propagated(t *testing.T) {
	stub := &schemaStub{err: errors.New("connection refused")}
	rules := []fetcher.SchemaRule{{Field: "version", Required: true, Type: "string"}}
	f, _ := fetcher.NewSchema(stub, rules)
	_, err := f.Fetch("svc", "http://example.com")
	if err == nil {
		t.Fatal("expected inner error to propagate")
	}
}

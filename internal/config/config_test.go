package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "driftwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	content := `
version: "1"
services:
  - name: api
    repository: https://github.com/example/api
    branch: main
    manifest: deploy/api.yaml
    labels:
      env: production
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cfg.Services))
	}
	svc := cfg.Services[0]
	if svc.Name != "api" {
		t.Errorf("expected name %q, got %q", "api", svc.Name)
	}
	if svc.Labels["env"] != "production" {
		t.Errorf("expected label env=production, got %q", svc.Labels["env"])
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_NoServices(t *testing.T) {
	path := writeTemp(t, "version: \"1\"\nservices: []\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty services")
	}
}

func TestLoad_DuplicateServiceName(t *testing.T) {
	content := `
version: "1"
services:
  - name: api
    repository: https://github.com/example/api
    manifest: deploy/api.yaml
  - name: api
    repository: https://github.com/example/api2
    manifest: deploy/api2.yaml
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate service name")
	}
}

func TestLoad_MissingManifest(t *testing.T) {
	content := `
version: "1"
services:
  - name: api
    repository: https://github.com/example/api
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

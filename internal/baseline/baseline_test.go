package baseline_test

import (
	"errors"
	"os"
	"testing"

	"github.com/yourorg/driftwatch/internal/baseline"
)

func makeEntry(name string) baseline.Entry {
	return baseline.Entry{
		ServiceName: name,
		Fields: map[string]interface{}{
			"version": "1.2.3",
			"replicas": float64(3),
		},
	}
}

func TestSave_And_Load(t *testing.T) {
	dir := t.TempDir()
	store, err := baseline.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	entry := makeEntry("svc-a")
	if err := store.Save(entry); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := store.Load("svc-a")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ServiceName != "svc-a" {
		t.Errorf("expected svc-a, got %s", loaded.ServiceName)
	}
	if loaded.Fields["version"] != "1.2.3" {
		t.Errorf("unexpected version: %v", loaded.Fields["version"])
	}
}

func TestLoad_Miss(t *testing.T) {
	dir := t.TempDir()
	store, _ := baseline.New(dir)
	_, err := store.Load("nonexistent")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected ErrNotExist, got %v", err)
	}
}

func TestDelete_RemovesBaseline(t *testing.T) {
	dir := t.TempDir()
	store, _ := baseline.New(dir)
	_ = store.Save(makeEntry("svc-b"))
	if err := store.Delete("svc-b"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.Load("svc-b")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected ErrNotExist after delete, got %v", err)
	}
}

func TestDelete_NonExistent_NoError(t *testing.T) {
	dir := t.TempDir()
	store, _ := baseline.New(dir)
	if err := store.Delete("ghost"); err != nil {
		t.Errorf("expected no error deleting nonexistent baseline, got %v", err)
	}
}

func TestList_ReturnsAll(t *testing.T) {
	dir := t.TempDir()
	store, _ := baseline.New(dir)
	_ = store.Save(makeEntry("svc-x"))
	_ = store.Save(makeEntry("svc-y"))
	entries, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestNew_CreatesDir(t *testing.T) {
	dir := t.TempDir() + "/nested/baseline"
	_, err := baseline.New(dir)
	if err != nil {
		t.Fatalf("New with nested dir: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

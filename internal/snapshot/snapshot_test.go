package snapshot_test

import (
	"os"
	"testing"
	"time"

	"github.com/driftwatch/internal/snapshot"
)

func TestSave_And_Load(t *testing.T) {
	dir := t.TempDir()
	store, err := snapshot.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	snap := snapshot.Snapshot{
		ServiceName: "api",
		Fields:      map[string]interface{}{"replicas": float64(3), "image": "nginx:1.25"},
	}

	if err := store.Save(snap); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load("api")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.ServiceName != "api" {
		t.Errorf("ServiceName = %q, want %q", loaded.ServiceName, "api")
	}
	if loaded.CapturedAt.IsZero() {
		t.Error("CapturedAt should not be zero")
	}
	if loaded.Fields["image"] != "nginx:1.25" {
		t.Errorf("Fields[image] = %v, want nginx:1.25", loaded.Fields["image"])
	}
}

func TestLoad_Miss(t *testing.T) {
	dir := t.TempDir()
	store, _ := snapshot.New(dir)

	_, err := store.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing snapshot, got nil")
	}
	if !os.IsNotExist(err) {
		// wrapped error — just check it's non-nil (already asserted above)
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	store, _ := snapshot.New(dir)

	first := snapshot.Snapshot{ServiceName: "svc", Fields: map[string]interface{}{"replicas": float64(1)}}
	second := snapshot.Snapshot{ServiceName: "svc", Fields: map[string]interface{}{"replicas": float64(5)}}

	_ = store.Save(first)
	time.Sleep(2 * time.Millisecond)
	_ = store.Save(second)

	loaded, err := store.Load("svc")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Fields["replicas"] != float64(5) {
		t.Errorf("replicas = %v, want 5", loaded.Fields["replicas"])
	}
}

func TestDelete_RemovesFile(t *testing.T) {
	dir := t.TempDir()
	store, _ := snapshot.New(dir)

	snap := snapshot.Snapshot{ServiceName: "del-svc", Fields: map[string]interface{}{}}
	_ = store.Save(snap)

	if err := store.Delete("del-svc"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := store.Load("del-svc"); err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestDelete_NonExistent_NoError(t *testing.T) {
	dir := t.TempDir()
	store, _ := snapshot.New(dir)

	if err := store.Delete("ghost"); err != nil {
		t.Errorf("Delete non-existent: unexpected error: %v", err)
	}
}

func TestNew_CreatesDir(t *testing.T) {
	dir := t.TempDir() + "/nested/snapshots"
	_, err := snapshot.New(dir)
	if err != nil {
		t.Fatalf("New with nested dir: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

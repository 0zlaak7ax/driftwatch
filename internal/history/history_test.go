package history_test

import (
	"os"
	"testing"

	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/history"
)

func makeResults(drifted bool) []drift.Result {
	return []drift.Result{
		{
			ServiceName: "svc-a",
			Drifted:     drifted,
			Diffs:       nil,
		},
	}
}

func TestRecord_And_List(t *testing.T) {
	dir := t.TempDir()
	store, err := history.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	results := makeResults(true)
	if err := store.Record(results); err != nil {
		t.Fatalf("Record: %v", err)
	}

	entries, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Total != 1 {
		t.Errorf("expected Total=1, got %d", entries[0].Total)
	}
	if entries[0].Drifted != 1 {
		t.Errorf("expected Drifted=1, got %d", entries[0].Drifted)
	}
}

func TestList_Empty(t *testing.T) {
	dir := t.TempDir()
	store, err := history.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	entries, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestRecord_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	store, err := history.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := store.Record(makeResults(false)); err != nil {
			t.Fatalf("Record[%d]: %v", i, err)
		}
	}

	entries, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestNew_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := base + "/nested/history"
	_, err := history.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

package audit_test

import (
	"os"
	"path/filepath"
	"testing"

	"driftwatch/internal/audit"
)

func TestRecord_And_List(t *testing.T) {
	dir := t.TempDir()
	log, err := audit.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := log.Record(audit.EventDetect, "svc-a", "drift detected", map[string]string{"fields": "2"}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	events, err := log.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Kind != audit.EventDetect {
		t.Errorf("expected kind %q, got %q", audit.EventDetect, events[0].Kind)
	}
	if events[0].Service != "svc-a" {
		t.Errorf("expected service svc-a, got %q", events[0].Service)
	}
	if events[0].Meta["fields"] != "2" {
		t.Errorf("expected meta fields=2, got %q", events[0].Meta["fields"])
	}
}

func TestList_Empty(t *testing.T) {
	dir := t.TempDir()
	log, _ := audit.New(dir)
	events, err := log.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestRecord_MultipleEvents(t *testing.T) {
	dir := t.TempDir()
	log, _ := audit.New(dir)
	kinds := []audit.EventKind{audit.EventDetect, audit.EventBaseline, audit.EventAlert}
	for _, k := range kinds {
		if err := log.Record(k, "svc", "msg", nil); err != nil {
			t.Fatalf("Record(%s): %v", k, err)
		}
	}
	events, err := log.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	for i, k := range kinds {
		if events[i].Kind != k {
			t.Errorf("event[%d]: expected %q, got %q", i, k, events[i].Kind)
		}
	}
}

func TestNew_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "nested", "audit")
	if _, err := audit.New(dir); err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected dir to exist: %v", err)
	}
}

func TestRecord_NoMeta(t *testing.T) {
	dir := t.TempDir()
	log, _ := audit.New(dir)
	if err := log.Record(audit.EventSnapshot, "svc-b", "snapshot saved", nil); err != nil {
		t.Fatalf("Record: %v", err)
	}
	events, _ := log.List()
	if len(events) != 1 {
		t.Fatalf("expected 1 event")
	}
	if events[0].Meta != nil {
		t.Errorf("expected nil meta, got %v", events[0].Meta)
	}
}

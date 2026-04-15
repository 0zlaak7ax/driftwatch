package audit_test

import (
	"testing"

	"driftwatch/internal/audit"
	"driftwatch/internal/drift"
)

func makeDriftResults(drifted bool) []drift.Result {
	return []drift.Result{
		{
			Service: "svc-x",
			Drifted: drifted,
			Deltas: func() []drift.Delta {
				if drifted {
					return []drift.Delta{{Field: "version", Expected: "1.0", Actual: "1.1"}}
				}
				return nil
			}(),
		},
	}
}

func TestDetectHook_NoDrift(t *testing.T) {
	dir := t.TempDir()
	log, _ := audit.New(dir)
	results := makeDriftResults(false)
	if err := audit.DetectHook(log, results); err != nil {
		t.Fatalf("DetectHook: %v", err)
	}
	events, _ := log.List()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Message != "all services in sync" {
		t.Errorf("unexpected message: %q", events[0].Message)
	}
}

func TestDetectHook_WithDrift(t *testing.T) {
	dir := t.TempDir()
	log, _ := audit.New(dir)
	results := makeDriftResults(true)
	if err := audit.DetectHook(log, results); err != nil {
		t.Fatalf("DetectHook: %v", err)
	}
	events, _ := log.List()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Service != "svc-x" {
		t.Errorf("expected service svc-x, got %q", events[0].Service)
	}
	if events[0].Meta["drifted_fields"] != "1" {
		t.Errorf("expected drifted_fields=1, got %q", events[0].Meta["drifted_fields"])
	}
}

func TestDetectHook_NilLog(t *testing.T) {
	if err := audit.DetectHook(nil, makeDriftResults(true)); err != nil {
		t.Errorf("expected no error with nil log, got %v", err)
	}
}

func TestBaselineHook_SaveAndDelete(t *testing.T) {
	dir := t.TempDir()
	log, _ := audit.New(dir)
	if err := audit.BaselineHook(log, "svc-y", "saved"); err != nil {
		t.Fatalf("BaselineHook saved: %v", err)
	}
	if err := audit.BaselineHook(log, "svc-y", "deleted"); err != nil {
		t.Fatalf("BaselineHook deleted: %v", err)
	}
	events, _ := log.List()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Meta["action"] != "saved" {
		t.Errorf("expected action=saved, got %q", events[0].Meta["action"])
	}
	if events[1].Meta["action"] != "deleted" {
		t.Errorf("expected action=deleted, got %q", events[1].Meta["action"])
	}
}

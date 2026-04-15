package alert_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/alert"
	"github.com/driftwatch/internal/drift"
)

func driftedResult(name string, fields ...string) drift.Result {
	return drift.Result{
		ServiceName:   name,
		Drifted:       true,
		DriftedFields: fields,
	}
}

func syncedResult(name string) drift.Result {
	return drift.Result{ServiceName: name, Drifted: false}
}

func TestEvaluate_NoDrift(t *testing.T) {
	a := alert.New(&bytes.Buffer{}, nil)
	results := []drift.Result{syncedResult("svc-a"), syncedResult("svc-b")}
	alerts := a.Evaluate(results)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestEvaluate_WarningLevel(t *testing.T) {
	a := alert.New(&bytes.Buffer{}, []string{"image"})
	results := []drift.Result{driftedResult("svc-a", "replicas")}
	alerts := a.Evaluate(results)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != alert.LevelWarning {
		t.Errorf("expected warning, got %s", alerts[0].Level)
	}
}

func TestEvaluate_CriticalLevel(t *testing.T) {
	a := alert.New(&bytes.Buffer{}, []string{"image"})
	results := []drift.Result{driftedResult("svc-b", "replicas", "image")}
	alerts := a.Evaluate(results)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != alert.LevelCritical {
		t.Errorf("expected critical, got %s", alerts[0].Level)
	}
}

func TestEmit_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	a := alert.New(&buf, []string{"image"})
	results := []drift.Result{
		driftedResult("svc-a", "replicas"),
		driftedResult("svc-b", "image"),
	}
	alerts := a.Evaluate(results)
	a.Emit(alerts)
	out := buf.String()
	if !strings.Contains(out, "[WARNING]") {
		t.Errorf("expected WARNING in output, got: %s", out)
	}
	if !strings.Contains(out, "[CRITICAL]") {
		t.Errorf("expected CRITICAL in output, got: %s", out)
	}
	if !strings.Contains(out, "svc-a") {
		t.Errorf("expected svc-a in output")
	}
}

func TestEvaluate_CriticalFieldCaseInsensitive(t *testing.T) {
	a := alert.New(&bytes.Buffer{}, []string{"IMAGE"})
	results := []drift.Result{driftedResult("svc-c", "image")}
	alerts := a.Evaluate(results)
	if alerts[0].Level != alert.LevelCritical {
		t.Errorf("expected critical for case-insensitive match, got %s", alerts[0].Level)
	}
}

// TestEvaluate_EmptyResults verifies that evaluating an empty result set
// produces no alerts and does not panic.
func TestEvaluate_EmptyResults(t *testing.T) {
	a := alert.New(&bytes.Buffer{}, []string{"image"})
	alerts := a.Evaluate([]drift.Result{})
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts for empty results, got %d", len(alerts))
	}
}

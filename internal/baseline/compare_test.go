package baseline_test

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/baseline"
)

func baseEntry() baseline.Entry {
	return baseline.Entry{
		ServiceName: "api",
		Fields: map[string]interface{}{
			"version":  "2.0.0",
			"replicas": float64(3),
			"region":   "us-east-1",
		},
	}
}

func TestCompare_InSync(t *testing.T) {
	live := map[string]interface{}{
		"version":  "2.0.0",
		"replicas": float64(3),
		"region":   "us-east-1",
	}
	res := baseline.Compare(baseEntry(), live)
	if !res.InSync {
		t.Errorf("expected in-sync, got deviations: %v", res.Deviations)
	}
	if len(res.Deviations) != 0 {
		t.Errorf("expected 0 deviations, got %d", len(res.Deviations))
	}
}

func TestCompare_ValueDrift(t *testing.T) {
	live := map[string]interface{}{
		"version":  "3.0.0",
		"replicas": float64(3),
		"region":   "us-east-1",
	}
	res := baseline.Compare(baseEntry(), live)
	if res.InSync {
		t.Error("expected drift, got in-sync")
	}
	if len(res.Deviations) != 1 {
		t.Fatalf("expected 1 deviation, got %d", len(res.Deviations))
	}
	if res.Deviations[0].Field != "version" {
		t.Errorf("expected deviation on 'version', got %s", res.Deviations[0].Field)
	}
}

func TestCompare_MissingLiveField(t *testing.T) {
	live := map[string]interface{}{
		"version": "2.0.0",
		// replicas and region missing
	}
	res := baseline.Compare(baseEntry(), live)
	if res.InSync {
		t.Error("expected drift due to missing fields")
	}
	if len(res.Deviations) != 2 {
		t.Errorf("expected 2 deviations, got %d", len(res.Deviations))
	}
}

func TestCompare_ExtraLiveField_Ignored(t *testing.T) {
	live := map[string]interface{}{
		"version":    "2.0.0",
		"replicas":   float64(3),
		"region":     "us-east-1",
		"extra_field": "ignored",
	}
	res := baseline.Compare(baseEntry(), live)
	if !res.InSync {
		t.Errorf("extra live fields should be ignored, got deviations: %v", res.Deviations)
	}
}

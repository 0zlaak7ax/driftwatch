package metrics_test

import (
	"testing"
	"time"

	"driftwatch/internal/metrics"
)

func makeRun(drifted, total, fields, fieldsDrifted int) metrics.RunMetrics {
	now := time.Now()
	return metrics.RunMetrics{
		StartedAt:       now,
		FinishedAt:      now.Add(200 * time.Millisecond),
		ServicesTotal:   total,
		ServicesDrifted: drifted,
		ServicesInSync:  total - drifted,
		FieldsChecked:   fields,
		FieldsDrifted:   fieldsDrifted,
	}
}

func TestRecord_And_Latest(t *testing.T) {
	c := metrics.New()
	r := makeRun(1, 3, 12, 2)
	c.Record(r)

	got, ok := c.Latest()
	if !ok {
		t.Fatal("expected Latest to return true")
	}
	if got.ServicesDrifted != 1 {
		t.Errorf("ServicesDrifted: want 1, got %d", got.ServicesDrifted)
	}
}

func TestLatest_Empty(t *testing.T) {
	c := metrics.New()
	_, ok := c.Latest()
	if ok {
		t.Fatal("expected Latest to return false on empty collector")
	}
}

func TestAll_ReturnsInOrder(t *testing.T) {
	c := metrics.New()
	c.Record(makeRun(0, 2, 8, 0))
	c.Record(makeRun(1, 2, 8, 3))
	c.Record(makeRun(2, 2, 8, 5))

	all := c.All()
	if len(all) != 3 {
		t.Fatalf("want 3 runs, got %d", len(all))
	}
	if all[0].ServicesDrifted != 0 || all[1].ServicesDrifted != 1 || all[2].ServicesDrifted != 2 {
		t.Error("runs not in insertion order")
	}
}

func TestReset_ClearsRuns(t *testing.T) {
	c := metrics.New()
	c.Record(makeRun(1, 4, 16, 2))
	c.Reset()

	if all := c.All(); len(all) != 0 {
		t.Errorf("expected 0 runs after Reset, got %d", len(all))
	}
}

func TestDuration(t *testing.T) {
	r := makeRun(0, 1, 4, 0)
	if r.Duration() < 100*time.Millisecond {
		t.Errorf("Duration too short: %v", r.Duration())
	}
}

func TestDriftRate(t *testing.T) {
	tests := []struct {
		drifted, total int
		want           float64
	}{
		{0, 0, 0},
		{0, 4, 0},
		{2, 4, 0.5},
		{4, 4, 1},
	}
	for _, tt := range tests {
		r := makeRun(tt.drifted, tt.total, 0, 0)
		if got := r.DriftRate(); got != tt.want {
			t.Errorf("DriftRate(%d/%d): want %.2f, got %.2f", tt.drifted, tt.total, tt.want, got)
		}
	}
}

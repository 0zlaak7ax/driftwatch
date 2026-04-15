package filter_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/filter"
)

func makeResults() []drift.Result {
	return []drift.Result{
		{
			ServiceName: "api",
			Drifted:     true,
			Tags:        map[string]string{"env": "prod", "team": "platform"},
		},
		{
			ServiceName: "worker",
			Drifted:     false,
			Tags:        map[string]string{"env": "prod", "team": "data"},
		},
		{
			ServiceName: "gateway",
			Drifted:     true,
			Tags:        map[string]string{"env": "staging", "team": "platform"},
		},
	}
}

func TestApply_NoOptions_ReturnsAll(t *testing.T) {
	results := makeResults()
	got := filter.Apply(results, filter.Options{})
	if len(got) != len(results) {
		t.Fatalf("expected %d results, got %d", len(results), len(got))
	}
}

func TestApply_OnlyDrifted(t *testing.T) {
	got := filter.Apply(makeResults(), filter.Options{OnlyDrifted: true})
	for _, r := range got {
		if !r.Drifted {
			t.Errorf("expected only drifted results, got in-sync service %q", r.ServiceName)
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 drifted results, got %d", len(got))
	}
}

func TestApply_FilterByServiceName(t *testing.T) {
	got := filter.Apply(makeResults(), filter.Options{Services: []string{"api", "worker"}})
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	for _, r := range got {
		if r.ServiceName == "gateway" {
			t.Error("gateway should have been excluded")
		}
	}
}

func TestApply_FilterByTag(t *testing.T) {
	got := filter.Apply(makeResults(), filter.Options{
		Tags: map[string]string{"team": "platform"},
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
}

func TestApply_FilterByTag_CaseInsensitive(t *testing.T) {
	got := filter.Apply(makeResults(), filter.Options{
		Tags: map[string]string{"env": "PROD"},
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 prod results, got %d", len(got))
	}
}

func TestApply_CombinedFilters(t *testing.T) {
	got := filter.Apply(makeResults(), filter.Options{
		OnlyDrifted: true,
		Tags:        map[string]string{"env": "prod"},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].ServiceName != "api" {
		t.Errorf("expected api, got %q", got[0].ServiceName)
	}
}

func TestApply_NoMatch_ReturnsEmpty(t *testing.T) {
	got := filter.Apply(makeResults(), filter.Options{Services: []string{"nonexistent"}})
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d", len(got))
	}
}

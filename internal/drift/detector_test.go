package drift_test

import (
	"errors"
	"testing"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/drift"
)

// mockFetcher implements drift.Fetcher for testing.
type mockFetcher struct {
	data map[string]map[string]string
	err  error
}

func (m *mockFetcher) Fetch(name string) (map[string]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[name], nil
}

func makeConfig(services []config.Service) *config.Config {
	return &config.Config{Services: services}
}

func TestDetect_InSync(t *testing.T) {
	cfg := makeConfig([]config.Service{
		{Name: "api", Params: map[string]string{"replicas": "3", "image": "api:v1"}},
	})
	fetcher := &mockFetcher{
		data: map[string]map[string]string{
			"api": {"replicas": "3", "image": "api:v1"},
		},
	}

	results, err := drift.New(fetcher).Detect(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != drift.StatusInSync {
		t.Errorf("expected in_sync, got %s", results[0].Status)
	}
	if len(results[0].Diffs) != 0 {
		t.Errorf("expected no diffs, got %v", results[0].Diffs)
	}
}

func TestDetect_Drifted(t *testing.T) {
	cfg := makeConfig([]config.Service{
		{Name: "worker", Params: map[string]string{"replicas": "5", "image": "worker:v2"}},
	})
	fetcher := &mockFetcher{
		data: map[string]map[string]string{
			"worker": {"replicas": "2", "image": "worker:v2"},
		},
	}

	results, _ := drift.New(fetcher).Detect(cfg)
	if results[0].Status != drift.StatusDrifted {
		t.Errorf("expected drifted, got %s", results[0].Status)
	}
	if len(results[0].Diffs) != 1 {
		t.Errorf("expected 1 diff, got %d", len(results[0].Diffs))
	}
	if results[0].Diffs[0].Field != "replicas" {
		t.Errorf("unexpected diff field: %s", results[0].Diffs[0].Field)
	}
}

func TestDetect_MissingField(t *testing.T) {
	cfg := makeConfig([]config.Service{
		{Name: "db", Params: map[string]string{"port": "5432"}},
	})
	fetcher := &mockFetcher{
		data: map[string]map[string]string{
			"db": {},
		},
	}

	results, _ := drift.New(fetcher).Detect(cfg)
	if results[0].Status != drift.StatusDrifted {
		t.Errorf("expected drifted for missing field")
	}
	if results[0].Diffs[0].Actual != "<missing>" {
		t.Errorf("expected <missing>, got %s", results[0].Diffs[0].Actual)
	}
}

func TestDetect_FetchError(t *testing.T) {
	cfg := makeConfig([]config.Service{
		{Name: "cache", Params: map[string]string{"maxmemory": "256mb"}},
	})
	fetcher := &mockFetcher{err: errors.New("connection refused")}

	results, _ := drift.New(fetcher).Detect(cfg)
	if results[0].Status != drift.StatusUnknown {
		t.Errorf("expected unknown on fetch error, got %s", results[0].Status)
	}
}

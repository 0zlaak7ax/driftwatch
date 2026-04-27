package runner

import (
	"errors"
	"fmt"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/fetcher"
)

// PriorityRunner runs drift detection where each service can specify an
// ordered list of fetch endpoints. The first successful endpoint wins.
type PriorityRunner struct {
	cfg     *config.Config
	detect  *drift.Detector
	workers int
}

// NewPriorityRunner creates a PriorityRunner.
// workers controls how many services are processed concurrently (min 1).
func NewPriorityRunner(cfg *config.Config, workers int) (*PriorityRunner, error) {
	if cfg == nil {
		return nil, errors.New("priority runner: config must not be nil")
	}
	if workers < 1 {
		workers = 1
	}
	d := drift.New(nil) // detector with no fetcher; fetching is handled per-service
	return &PriorityRunner{cfg: cfg, detect: d, workers: workers}, nil
}

// BuildFetcher constructs a priority fetcher for a service whose config
// contains multiple endpoint URLs ordered by preference.
func BuildFetcher(urls []string, timeout int) (fetcher.Fetcher, error) {
	if len(urls) == 0 {
		return nil, errors.New("priority runner: at least one URL required")
	}
	entries := make([]fetcher.PriorityEntry, len(urls))
	base := fetcher.New(0)
	for i, u := range urls {
		if u == "" {
			return nil, fmt.Errorf("priority runner: URL at index %d is empty", i)
		}
		entries[i] = fetcher.PriorityEntry{
			Fetcher:  base,
			Priority: i,
		}
		_ = u // URL is passed to Fetch at call-time, not construction-time
	}
	return fetcher.NewPriority(entries)
}

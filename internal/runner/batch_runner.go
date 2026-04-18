package runner

import (
	"context"
	"fmt"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/fetcher"
)

// BatchRunner uses BatchFetcher to pre-warm all service endpoints concurrently
// before the main detection run, reducing total wall-clock time.
type BatchRunner struct {
	cfg     *config.Config
	batcher *fetcher.BatchFetcher
}

// NewBatchRunner creates a BatchRunner for the given config and inner fetcher.
func NewBatchRunner(cfg *config.Config, inner fetcher.Fetcher, workers int) (*BatchRunner, error) {
	if cfg == nil {
		return nil, fmt.Errorf("batch runner: config must not be nil")
	}
	bf, err := fetcher.NewBatch(inner, workers)
	if err != nil {
		return nil, fmt.Errorf("batch runner: %w", err)
	}
	return &BatchRunner{cfg: cfg, batcher: bf}, nil
}

// Prefetch fetches all configured service endpoints and returns any errors keyed by service name.
func (br *BatchRunner) Prefetch(ctx context.Context) map[string]error {
	urls := make([]string, len(br.cfg.Services))
	names := make([]string, len(br.cfg.Services))
	for i, svc := range br.cfg.Services {
		urls[i] = svc.URL
		names[i] = svc.Name
	}

	results := br.batcher.FetchAll(ctx, urls)
	errs := make(map[string]error)
	for i, r := range results {
		if r.Err != nil {
			errs[names[i]] = r.Err
		}
	}
	return errs
}

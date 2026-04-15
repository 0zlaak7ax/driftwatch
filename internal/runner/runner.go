package runner

import (
	"context"
	"fmt"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/fetcher"
	"github.com/driftwatch/internal/output"
)

// Runner orchestrates the full drift detection pipeline.
type Runner struct {
	cfg       *config.Config
	formatter *output.Formatter
}

// New creates a Runner from the provided config and output format string.
func New(cfg *config.Config, format string) (*Runner, error) {
	fmt_, err := output.ParseFormat(format)
	if err != nil {
		return nil, fmt.Errorf("runner: invalid format: %w", err)
	}
	f := output.New(fmt_)
	return &Runner{cfg: cfg, formatter: f}, nil
}

// Run executes drift detection for all configured services and writes the
// formatted report to stdout. It returns true when drift is detected.
func (r *Runner) Run(ctx context.Context) (bool, error) {
	results := make([]drift.Result, 0, len(r.cfg.Services))

	for _, svc := range r.cfg.Services {
		f, err := fetcher.New(svc.URL, svc.TimeoutSeconds)
		if err != nil {
			return false, fmt.Errorf("runner: service %q: %w", svc.Name, err)
		}

		detector := drift.New(f)
		res, err := detector.Detect(ctx, svc)
		if err != nil {
			return false, fmt.Errorf("runner: service %q: %w", svc.Name, err)
		}
		results = append(results, res)
	}

	if err := r.formatter.Print(results); err != nil {
		return false, fmt.Errorf("runner: formatting output: %w", err)
	}

	for _, res := range results {
		if res.Drifted {
			return true, nil
		}
	}
	return false, nil
}

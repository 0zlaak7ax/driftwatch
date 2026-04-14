package drift

import (
	"fmt"

	"github.com/driftwatch/internal/config"
)

// Status represents the drift state of a service.
type Status string

const (
	StatusInSync  Status = "in_sync"
	StatusDrifted Status = "drifted"
	StatusUnknown Status = "unknown"
)

// Result holds the drift detection result for a single service.
type Result struct {
	ServiceName string
	Status      Status
	Diffs       []Diff
}

// Diff describes a single configuration discrepancy.
type Diff struct {
	Field    string
	Expected string
	Actual   string
}

// Fetcher retrieves the live configuration for a service by name.
type Fetcher interface {
	Fetch(serviceName string) (map[string]string, error)
}

// Detector compares declared config against live state.
type Detector struct {
	fetcher Fetcher
}

// New creates a new Detector with the given Fetcher.
func New(fetcher Fetcher) *Detector {
	return &Detector{fetcher: fetcher}
}

// Detect runs drift detection for all services defined in cfg.
func (d *Detector) Detect(cfg *config.Config) ([]Result, error) {
	results := make([]Result, 0, len(cfg.Services))

	for _, svc := range cfg.Services {
		live, err := d.fetcher.Fetch(svc.Name)
		if err != nil {
			results = append(results, Result{
				ServiceName: svc.Name,
				Status:      StatusUnknown,
				Diffs:       nil,
			})
			continue
		}

		diffs := compare(svc.Params, live)
		status := StatusInSync
		if len(diffs) > 0 {
			status = StatusDrifted
		}

		results = append(results, Result{
			ServiceName: svc.Name,
			Status:      status,
			Diffs:       diffs,
		})
	}

	return results, nil
}

// compare returns the list of differences between declared and live params.
func compare(declared, live map[string]string) []Diff {
	var diffs []Diff

	for key, expectedVal := range declared {
		actualVal, ok := live[key]
		if !ok {
			diffs = append(diffs, Diff{
				Field:    key,
				Expected: expectedVal,
				Actual:   fmt.Sprintf("<missing>"),
			})
			continue
		}
		if actualVal != expectedVal {
			diffs = append(diffs, Diff{
				Field:    key,
				Expected: expectedVal,
				Actual:   actualVal,
			})
		}
	}

	return diffs
}

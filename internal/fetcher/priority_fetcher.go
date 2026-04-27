package fetcher

import (
	"errors"
	"fmt"
	"sort"
)

// PriorityEntry pairs a fetcher with a numeric priority (lower = higher priority).
type PriorityEntry struct {
	Fetcher  Fetcher
	Priority int
}

// priorityFetcher tries each inner fetcher in priority order, returning the
// first successful result. All errors are collected and returned only when
// every fetcher has failed.
type priorityFetcher struct {
	entries []PriorityEntry
}

// NewPriority creates a priorityFetcher from the provided entries.
// Entries are sorted ascending by Priority so the lowest value runs first.
// Returns an error if entries is empty or any Fetcher is nil.
func NewPriority(entries []PriorityEntry) (Fetcher, error) {
	if len(entries) == 0 {
		return nil, errors.New("priority fetcher: at least one entry required")
	}
	for i, e := range entries {
		if e.Fetcher == nil {
			return nil, fmt.Errorf("priority fetcher: entry %d has nil fetcher", i)
		}
	}
	sorted := make([]PriorityEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})
	return &priorityFetcher{entries: sorted}, nil
}

// Fetch iterates entries in priority order and returns the first success.
func (p *priorityFetcher) Fetch(url string) (map[string]interface{}, error) {
	var errs []error
	for _, e := range p.entries {
		result, err := e.Fetcher.Fetch(url)
		if err == nil {
			return result, nil
		}
		errs = append(errs, fmt.Errorf("priority %d: %w", e.Priority, err))
	}
	return nil, fmt.Errorf("priority fetcher: all fetchers failed: %w", errors.Join(errs...))
}

package baseline

import (
	"fmt"
	"time"
)

// Fetcher retrieves the current live fields for a service by name.
type Fetcher interface {
	Fetch(serviceName, url string) (map[string]interface{}, error)
}

// Manager orchestrates capturing and comparing baselines.
type Manager struct {
	store   *Store
	fetcher Fetcher
}

// NewManager creates a Manager using the provided store and fetcher.
func NewManager(store *Store, fetcher Fetcher) *Manager {
	return &Manager{store: store, fetcher: fetcher}
}

// Capture fetches live state for a service and saves it as the baseline.
func (m *Manager) Capture(serviceName, url string) error {
	fields, err := m.fetcher.Fetch(serviceName, url)
	if err != nil {
		return fmt.Errorf("baseline capture %s: %w", serviceName, err)
	}
	return m.store.Save(Entry{
		ServiceName: serviceName,
		CapturedAt:  time.Now().UTC(),
		Fields:      fields,
	})
}

// Evaluate loads the baseline for a service, fetches live state, and compares them.
// Returns ErrNotExist (wrapped) if no baseline has been captured yet.
func (m *Manager) Evaluate(serviceName, url string) (CompareResult, error) {
	entry, err := m.store.Load(serviceName)
	if err != nil {
		return CompareResult{}, fmt.Errorf("baseline evaluate %s: %w", serviceName, err)
	}
	live, err := m.fetcher.Fetch(serviceName, url)
	if err != nil {
		return CompareResult{}, fmt.Errorf("baseline evaluate fetch %s: %w", serviceName, err)
	}
	return Compare(entry, live), nil
}

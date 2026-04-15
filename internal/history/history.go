package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

// Entry represents a single drift check recorded in history.
type Entry struct {
	Timestamp time.Time      `json:"timestamp"`
	Results   []drift.Result `json:"results"`
	Drifted   int            `json:"drifted"`
	Total     int            `json:"total"`
}

// Store manages persisting and loading drift history entries.
type Store struct {
	dir string
}

// New creates a new Store that persists entries under dir.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("history: create dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Record saves a drift run as a timestamped JSON file.
func (s *Store) Record(results []drift.Result) error {
	drifted := 0
	for _, r := range results {
		if r.Drifted {
			drifted++
		}
	}
	entry := Entry{
		Timestamp: time.Now().UTC(),
		Results:   results,
		Drifted:   drifted,
		Total:     len(results),
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}
	filename := entry.Timestamp.Format("20060102T150405Z") + ".json"
	path := filepath.Join(s.dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("history: write file: %w", err)
	}
	return nil
}

// List returns all recorded entries sorted oldest-first.
func (s *Store) List() ([]Entry, error) {
	matches, err := filepath.Glob(filepath.Join(s.dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("history: glob: %w", err)
	}
	var entries []Entry
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("history: read %s: %w", path, err)
		}
		var e Entry
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, fmt.Errorf("history: unmarshal %s: %w", path, err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

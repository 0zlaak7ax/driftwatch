package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a saved baseline for a service.
type Entry struct {
	ServiceName string                 `json:"service_name"`
	CapturedAt  time.Time              `json:"captured_at"`
	Fields      map[string]interface{} `json:"fields"`
}

// Store persists and retrieves baseline entries on disk.
type Store struct {
	dir string
}

// New creates a new Store rooted at dir, creating the directory if needed.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("baseline: create dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Save writes a baseline entry for the given service, overwriting any existing one.
func (s *Store) Save(entry Entry) error {
	entry.CapturedAt = time.Now().UTC()
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	return os.WriteFile(s.path(entry.ServiceName), data, 0o644)
}

// Load retrieves the baseline entry for the named service.
// Returns os.ErrNotExist if no baseline has been saved.
func (s *Store) Load(serviceName string) (Entry, error) {
	data, err := os.ReadFile(s.path(serviceName))
	if err != nil {
		return Entry{}, fmt.Errorf("baseline: read: %w", err)
	}
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return Entry{}, fmt.Errorf("baseline: unmarshal: %w", err)
	}
	return entry, nil
}

// Delete removes the baseline for the named service. It is not an error if the
// baseline does not exist.
func (s *Store) Delete(serviceName string) error {
	err := os.Remove(s.path(serviceName))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// List returns all stored baseline entries.
func (s *Store) List() ([]Entry, error) {
	glob := filepath.Join(s.dir, "*.json")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err != nil {
			return nil, err
		}
		var e Entry
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *Store) path(serviceName string) string {
	return filepath.Join(s.dir, serviceName+".json")
}

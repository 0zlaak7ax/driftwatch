package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot represents a point-in-time capture of a service's live configuration.
type Snapshot struct {
	ServiceName string                 `json:"service_name"`
	CapturedAt  time.Time              `json:"captured_at"`
	Fields      map[string]interface{} `json:"fields"`
}

// Store persists and retrieves snapshots from disk.
type Store struct {
	dir string
}

// New creates a new Store, ensuring the directory exists.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("snapshot: create dir %q: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Save writes a snapshot for the given service to disk.
func (s *Store) Save(snap Snapshot) error {
	snap.CapturedAt = time.Now().UTC()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal %q: %w", snap.ServiceName, err)
	}
	path := s.filePath(snap.ServiceName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("snapshot: write %q: %w", path, err)
	}
	return nil
}

// Load retrieves the latest snapshot for a service. Returns os.ErrNotExist if none found.
func (s *Store) Load(serviceName string) (Snapshot, error) {
	path := s.filePath(serviceName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: read %q: %w", path, err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: unmarshal %q: %w", serviceName, err)
	}
	return snap, nil
}

// Delete removes the snapshot file for a service.
func (s *Store) Delete(serviceName string) error {
	path := s.filePath(serviceName)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("snapshot: delete %q: %w", path, err)
	}
	return nil
}

func (s *Store) filePath(serviceName string) string {
	return filepath.Join(s.dir, serviceName+".snapshot.json")
}

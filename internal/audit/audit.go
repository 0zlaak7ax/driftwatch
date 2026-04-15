package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventKind describes the type of audit event.
type EventKind string

const (
	EventDetect   EventKind = "detect"
	EventBaseline EventKind = "baseline"
	EventAlert    EventKind = "alert"
	EventSnapshot EventKind = "snapshot"
)

// Event represents a single auditable action.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Kind      EventKind `json:"kind"`
	Service   string    `json:"service,omitempty"`
	Message   string    `json:"message"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Log is an append-only audit log backed by a JSONL file.
type Log struct {
	mu  sync.Mutex
	dir string
}

// New creates a new audit Log that writes to dir/audit.jsonl.
func New(dir string) (*Log, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("audit: create dir: %w", err)
	}
	return &Log{dir: dir}, nil
}

// Record appends an event to the audit log.
func (l *Log) Record(kind EventKind, service, message string, meta map[string]string) error {
	e := Event{
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		Service:   service,
		Message:   message,
		Meta:      meta,
	}
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := os.OpenFile(l.path(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("audit: open file: %w", err)
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// List returns all recorded events in order.
func (l *Log) List() ([]Event, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	data, err := os.ReadFile(l.path())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("audit: read file: %w", err)
	}
	var events []Event
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("audit: unmarshal: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

func (l *Log) path() string {
	return filepath.Join(l.dir, "audit.jsonl")
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

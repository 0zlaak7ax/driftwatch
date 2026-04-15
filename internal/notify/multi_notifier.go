package notify

import (
	"errors"
	"fmt"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Sender is the interface satisfied by Notifier and any custom channel.
type Sender interface {
	Notify(results []drift.Result) error
}

// Multi fans out notifications to multiple Senders, collecting all errors.
type Multi struct {
	senders []Sender
}

// NewMulti returns a Multi that dispatches to all provided senders.
func NewMulti(senders ...Sender) (*Multi, error) {
	for i, s := range senders {
		if s == nil {
			return nil, fmt.Errorf("notify: sender at index %d is nil", i)
		}
	}
	return &Multi{senders: senders}, nil
}

// Notify calls every registered sender and returns a combined error if any fail.
func (m *Multi) Notify(results []drift.Result) error {
	var errs []string
	for _, s := range m.senders {
		if err := s.Notify(results); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Count returns the number of registered senders.
func (m *Multi) Count() int {
	return len(m.senders)
}

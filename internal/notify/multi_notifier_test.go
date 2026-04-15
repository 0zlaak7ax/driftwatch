package notify_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/notify"
)

// fakeSender is a test double for notify.Sender.
type fakeSender struct {
	called bool
	err    error
}

func (f *fakeSender) Notify(_ []drift.Result) error {
	f.called = true
	return f.err
}

func TestNewMulti_NilSender_ReturnsError(t *testing.T) {
	_, err := notify.NewMulti(nil)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestMulti_Count(t *testing.T) {
	a, b := &fakeSender{}, &fakeSender{}
	m, err := notify.NewMulti(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Count() != 2 {
		t.Errorf("expected 2 senders, got %d", m.Count())
	}
}

func TestMulti_Notify_CallsAllSenders(t *testing.T) {
	a, b := &fakeSender{}, &fakeSender{}
	m, _ := notify.NewMulti(a, b)

	results := []drift.Result{{Service: "svc", Drifted: true}}
	if err := m.Notify(results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.called {
		t.Error("sender a was not called")
	}
	if !b.called {
		t.Error("sender b was not called")
	}
}

func TestMulti_Notify_CollectsErrors(t *testing.T) {
	a := &fakeSender{err: errors.New("err-a")}
	b := &fakeSender{err: errors.New("err-b")}
	m, _ := notify.NewMulti(a, b)

	err := m.Notify(nil)
	if err == nil {
		t.Fatal("expected combined error")
	}
	if !strings.Contains(err.Error(), "err-a") || !strings.Contains(err.Error(), "err-b") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMulti_Notify_PartialError_StillCallsAll(t *testing.T) {
	a := &fakeSender{err: errors.New("fail")}
	b := &fakeSender{}
	m, _ := notify.NewMulti(a, b)

	_ = m.Notify(nil)
	if !b.called {
		t.Error("sender b should be called even if sender a fails")
	}
}

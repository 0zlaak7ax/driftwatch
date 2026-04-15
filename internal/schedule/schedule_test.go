package schedule_test

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"driftwatch/internal/schedule"
)

type mockRunner struct {
	callCount atomic.Int32
	errOnCall int32 // if > 0, return error on that call number
}

func (m *mockRunner) Run(_ context.Context) error {
	call := m.callCount.Add(1)
	if m.errOnCall > 0 && call == m.errOnCall {
		return errors.New("mock run error")
	}
	return nil
}

func silentLogger() *log.Logger {
	return log.New(log.Writer(), "", 0)
}

func TestScheduler_RunsImmediately(t *testing.T) {
	runner := &mockRunner{}
	s := schedule.New(runner, 10*time.Second, silentLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	s.Start(ctx)

	if runner.callCount.Load() < 1 {
		t.Error("expected runner to be called at least once immediately")
	}
}

func TestScheduler_RunsOnInterval(t *testing.T) {
	runner := &mockRunner{}
	s := schedule.New(runner, 80*time.Millisecond, silentLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	s.Start(ctx)

	// Should have run: 1 immediate + ~2-3 ticks within 300ms
	count := runner.callCount.Load()
	if count < 2 {
		t.Errorf("expected at least 2 runs, got %d", count)
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	runner := &mockRunner{}
	s := schedule.New(runner, 50*time.Millisecond, silentLogger())

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	time.Sleep(120 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Error("scheduler did not stop after context cancellation")
	}
}

func TestScheduler_ContinuesOnRunError(t *testing.T) {
	runner := &mockRunner{errOnCall: 1}
	s := schedule.New(runner, 70*time.Millisecond, silentLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	s.Start(ctx)

	if runner.callCount.Load() < 2 {
		t.Error("expected scheduler to continue running after an error")
	}
}

func TestNew_NilLogger(t *testing.T) {
	runner := &mockRunner{}
	s := schedule.New(runner, time.Second, nil)
	if s == nil {
		t.Error("expected non-nil scheduler with nil logger")
	}
}

package schedule

import (
	"context"
	"log"
	"time"
)

// Runner defines the interface for anything that can be run on a schedule.
type Runner interface {
	Run(ctx context.Context) error
}

// Scheduler periodically executes a Runner at a fixed interval.
type Scheduler struct {
	runner   Runner
	interval time.Duration
	logger   *log.Logger
}

// New creates a new Scheduler with the given runner and interval.
func New(runner Runner, interval time.Duration, logger *log.Logger) *Scheduler {
	if logger == nil {
		logger = log.Default()
	}
	return &Scheduler{
		runner:   runner,
		interval: interval,
		logger:   logger,
	}
}

// Start begins the scheduling loop. It runs the runner immediately, then
// on each tick of the interval. It blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Printf("scheduler: starting with interval %s", s.interval)

	if err := s.runOnce(ctx); err != nil {
		s.logger.Printf("scheduler: run error: %v", err)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Println("scheduler: stopped")
			return
		case <-ticker.C:
			if err := s.runOnce(ctx); err != nil {
				s.logger.Printf("scheduler: run error: %v", err)
			}
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) error {
	s.logger.Println("scheduler: executing run")
	return s.runner.Run(ctx)
}

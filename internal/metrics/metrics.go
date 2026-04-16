package metrics

import (
	"sync"
	"time"
)

// RunMetrics holds statistics collected during a single drift-detection run.
type RunMetrics struct {
	StartedAt      time.Time
	FinishedAt     time.Time
	ServicesTotal  int
	ServicesDrifted int
	ServicesInSync  int
	FieldsChecked  int
	FieldsDrifted  int
}

// Duration returns how long the run took.
func (m RunMetrics) Duration() time.Duration {
	return m.FinishedAt.Sub(m.StartedAt)
}

// DriftRate returns the fraction of services that drifted (0–1).
func (m RunMetrics) DriftRate() float64 {
	if m.ServicesTotal == 0 {
		return 0
	}
	return float64(m.ServicesDrifted) / float64(m.ServicesTotal)
}

// Collector accumulates metrics across one or more runs.
type Collector struct {
	mu   sync.Mutex
	runs []RunMetrics
}

// New returns an initialised Collector.
func New() *Collector {
	return &Collector{}
}

// Record appends a completed RunMetrics snapshot to the collector.
func (c *Collector) Record(m RunMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs = append(c.runs, m)
}

// Latest returns the most-recently recorded RunMetrics and true, or a zero
// value and false when no runs have been recorded yet.
func (c *Collector) Latest() (RunMetrics, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.runs) == 0 {
		return RunMetrics{}, false
	}
	return c.runs[len(c.runs)-1], true
}

// All returns a copy of every recorded RunMetrics in insertion order.
func (c *Collector) All() []RunMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]RunMetrics, len(c.runs))
	copy(out, c.runs)
	return out
}

// Reset clears all recorded runs.
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs = nil
}

// Summary returns aggregate totals across all recorded runs.
// It returns false if no runs have been recorded yet.
func (c *Collector) Summary() (RunMetrics, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.runs) == 0 {
		return RunMetrics{}, false
	}
	var agg RunMetrics
	for _, r := range c.runs {
		agg.ServicesTotal += r.ServicesTotal
		agg.ServicesDrifted += r.ServicesDrifted
		agg.ServicesInSync += r.ServicesInSync
		agg.FieldsChecked += r.FieldsChecked
		agg.FieldsDrifted += r.FieldsDrifted
	}
	return agg, true
}

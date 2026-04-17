package fetcher

import (
	"errors"
	"time"

	"github.com/driftwatch/internal/metrics"
)

// MetricsFetcher wraps a Fetcher and records fetch metrics.
type MetricsFetcher struct {
	inner   Fetcher
	metrics *metrics.Store
	service string
}

// NewMetrics returns a MetricsFetcher wrapping inner.
func NewMetrics(inner Fetcher, store *metrics.Store, service string) (*MetricsFetcher, error) {
	if inner == nil {
		return nil, errors.New("metrics fetcher: inner fetcher must not be nil")
	}
	if store == nil {
		return nil, errors.New("metrics fetcher: store must not be nil")
	}
	if service == "" {
		return nil, errors.New("metrics fetcher: service name must not be empty")
	}
	return &MetricsFetcher{inner: inner, metrics: store, service: service}, nil
}

// Fetch delegates to the inner fetcher and records latency and success/failure.
func (m *MetricsFetcher) Fetch(url string) (map[string]any, error) {
	start := time.Now()
	result, err := m.inner.Fetch(url)
	elapsed := time.Since(start)

	run := metrics.FetchRun{
		Service:  m.service,
		Duration: elapsed,
		Success:  err == nil,
	}
	m.metrics.Record(run)

	return result, err
}

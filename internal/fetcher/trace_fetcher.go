package fetcher

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// TraceFetcher wraps a Fetcher and emits structured trace lines for each fetch,
// recording the service name, duration, and whether the call succeeded.
type TraceFetcher struct {
	inner   Fetcher
	writer  io.Writer
}

// NewTrace returns a TraceFetcher that writes trace output to w.
// If w is nil, os.Stderr is used.
// Returns an error if inner is nil.
func NewTrace(inner Fetcher, w io.Writer) (*TraceFetcher, error) {
	if inner == nil {
		return nil, fmt.Errorf("trace fetcher: inner fetcher must not be nil")
	}
	if w == nil {
		w = os.Stderr
	}
	return &TraceFetcher{inner: inner, writer: w}, nil
}

// Fetch delegates to the inner fetcher, measuring latency and writing a trace
// line in the format: [TRACE] service=<name> duration=<ms>ms ok=<bool>
func (t *TraceFetcher) Fetch(ctx context.Context, service, url string) (map[string]interface{}, error) {
	start := time.Now()
	result, err := t.inner.Fetch(ctx, service, url)
	elapsed := time.Since(start)

	ok := err == nil
	fmt.Fprintf(t.writer, "[TRACE] service=%s duration=%dms ok=%t\n",
		service, elapsed.Milliseconds(), ok)

	return result, err
}

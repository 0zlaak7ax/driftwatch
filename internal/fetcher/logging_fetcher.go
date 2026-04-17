package fetcher

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// LoggingFetcher wraps a Fetcher and logs each request with duration and status.
type LoggingFetcher struct {
	inner  Fetcher
	writer io.Writer
}

// NewLogging returns a LoggingFetcher wrapping inner. If writer is nil, os.Stderr is used.
func NewLogging(inner Fetcher, writer io.Writer) (*LoggingFetcher, error) {
	if inner == nil {
		return nil, errors.New("logging fetcher: inner fetcher must not be nil")
	}
	if writer == nil {
		writer = os.Stderr
	}
	return &LoggingFetcher{inner: inner, writer: writer}, nil
}

// Fetch executes the inner fetch, logs the outcome, and returns the result.
func (l *LoggingFetcher) Fetch(url string) (map[string]interface{}, error) {
	start := time.Now()
	result, err := l.inner.Fetch(url)
	duration := time.Since(start)

	if err != nil {
		fmt.Fprintf(l.writer, "[driftwatch] fetch error url=%s duration=%s err=%v\n", url, duration.Round(time.Millisecond), err)
		return nil, err
	}

	fmt.Fprintf(l.writer, "[driftwatch] fetch ok url=%s duration=%s fields=%d\n", url, duration.Round(time.Millisecond), len(result))
	return result, nil
}

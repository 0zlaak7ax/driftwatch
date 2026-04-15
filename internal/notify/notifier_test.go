package notify_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/notify"
)

func driftedResults() []drift.Result {
	return []drift.Result{
		{Service: "api", Drifted: true, Fields: []drift.FieldDiff{
			{Field: "version", Expected: "1.0", Actual: "1.1"},
		}},
		{Service: "worker", Drifted: false},
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	n, err := notify.New(notify.Config{Channel: notify.ChannelStdout, MinDrifted: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
}

func TestNew_NegativeMinDrifted_ReturnsError(t *testing.T) {
	_, err := notify.New(notify.Config{MinDrifted: -1})
	if err == nil {
		t.Fatal("expected error for negative MinDrifted")
	}
}

func TestNew_UnknownChannel_ReturnsError(t *testing.T) {
	_, err := notify.New(notify.Config{Channel: "slack", MinDrifted: 0})
	if err == nil {
		t.Fatal("expected error for unknown channel")
	}
}

func TestNotify_BelowThreshold_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	n, _ := notify.New(notify.Config{
		Channel:    notify.ChannelStdout,
		MinDrifted: 5,
		Writer:     &buf,
	})

	if err := n.Notify(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %q", buf.String())
	}
}

func TestNotify_AtThreshold_WritesMessage(t *testing.T) {
	var buf bytes.Buffer
	n, _ := notify.New(notify.Config{
		Channel:    notify.ChannelStdout,
		MinDrifted: 1,
		Writer:     &buf,
	})

	if err := n.Notify(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1/2 service(s) drifted") {
		t.Errorf("unexpected output: %q", out)
	}
	if !strings.Contains(out, "api") {
		t.Errorf("expected service name in output: %q", out)
	}
}

func TestNotify_WithPrefix(t *testing.T) {
	var buf bytes.Buffer
	n, _ := notify.New(notify.Config{
		Channel:    notify.ChannelStdout,
		MinDrifted: 1,
		Prefix:     "[ALERT]",
		Writer:     &buf,
	})

	_ = n.Notify(driftedResults())
	if !strings.HasPrefix(buf.String(), "[ALERT]") {
		t.Errorf("expected prefix in output: %q", buf.String())
	}
}

func TestNotify_NoDriftedServices_ZeroThreshold_WritesMessage(t *testing.T) {
	var buf bytes.Buffer
	n, _ := notify.New(notify.Config{
		Channel:    notify.ChannelStdout,
		MinDrifted: 0,
		Writer:     &buf,
	})

	syncedOnly := []drift.Result{{Service: "db", Drifted: false}}
	_ = n.Notify(syncedOnly)
	if !strings.Contains(buf.String(), "0/1") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

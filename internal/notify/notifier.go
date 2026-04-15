package notify

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Channel represents a notification delivery channel.
type Channel string

const (
	ChannelStdout  Channel = "stdout"
	ChannelStderr  Channel = "stderr"
	ChannelWebhook Channel = "webhook"
)

// Config holds configuration for a Notifier.
type Config struct {
	Channel    Channel
	MinDrifted int    // minimum number of drifted services to trigger notification
	Prefix     string // optional message prefix
	Writer     io.Writer
}

// Notifier sends drift notifications over a configured channel.
type Notifier struct {
	cfg Config
	out io.Writer
}

// New creates a new Notifier. If cfg.Writer is nil and channel is stdout/stderr,
// a default writer is selected.
func New(cfg Config) (*Notifier, error) {
	if cfg.MinDrifted < 0 {
		return nil, fmt.Errorf("notify: MinDrifted must be >= 0, got %d", cfg.MinDrifted)
	}

	out := cfg.Writer
	if out == nil {
		switch cfg.Channel {
		case ChannelStdout, "":
			out = os.Stdout
		case ChannelStderr:
			out = os.Stderr
		case ChannelWebhook:
			// webhook channel uses its own transport; writer unused
			out = io.Discard
		default:
			return nil, fmt.Errorf("notify: unknown channel %q", cfg.Channel)
		}
	}

	return &Notifier{cfg: cfg, out: out}, nil
}

// Notify writes a notification if the number of drifted results meets the
// threshold defined by cfg.MinDrifted.
func (n *Notifier) Notify(results []drift.Result) error {
	drifted := countDrifted(results)
	if drifted < n.cfg.MinDrifted {
		return nil
	}

	lines := buildMessage(n.cfg.Prefix, results)
	_, err := fmt.Fprintln(n.out, lines)
	return err
}

func countDrifted(results []drift.Result) int {
	count := 0
	for _, r := range results {
		if r.Drifted {
			count++
		}
	}
	return count
}

func buildMessage(prefix string, results []drift.Result) string {
	var sb strings.Builder
	if prefix != "" {
		sb.WriteString(prefix)
		sb.WriteString(" ")
	}
	drifted := countDrifted(results)
	sb.WriteString(fmt.Sprintf("[driftwatch] %d/%d service(s) drifted", drifted, len(results)))
	for _, r := range results {
		if r.Drifted {
			sb.WriteString(fmt.Sprintf("\n  - %s: %d field(s) changed", r.Service, len(r.Fields)))
		}
	}
	return sb.String()
}

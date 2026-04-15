package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Format represents the output format for drift reports.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Printer writes drift results to an output stream.
type Printer struct {
	w      io.Writer
	format Format
}

// New creates a new Printer that writes to w using the given format.
func New(w io.Writer, format Format) *Printer {
	return &Printer{w: w, format: format}
}

// Print writes all drift results to the configured output.
func (p *Printer) Print(results []drift.Result) error {
	switch p.format {
	case FormatJSON:
		return p.printJSON(results)
	default:
		return p.printText(results)
	}
}

func (p *Printer) printText(results []drift.Result) error {
	if len(results) == 0 {
		_, err := fmt.Fprintln(p.w, "✓ No drift detected. All services are in sync.")
		return err
	}

	for _, r := range results {
		if r.InSync {
			fmt.Fprintf(p.w, "✓ %s: in sync\n", r.ServiceName)
			continue
		}
		fmt.Fprintf(p.w, "✗ %s: drift detected\n", r.ServiceName)
		for _, d := range r.Diffs {
			fmt.Fprintf(p.w, "    field=%s expected=%q actual=%q\n",
				d.Field, d.Expected, d.Actual)
		}
	}
	return nil
}

func (p *Printer) printJSON(results []drift.Result) error {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("  {\"service\": %q, \"in_sync\": %v, \"diffs\": [", r.ServiceName, r.InSync))
		for j, d := range r.Diffs {
			sb.WriteString(fmt.Sprintf("{\"field\": %q, \"expected\": %q, \"actual\": %q}",
				d.Field, d.Expected, d.Actual))
			if j < len(r.Diffs)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("]}")
		if i < len(results)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("]\n")
	_, err := fmt.Fprint(p.w, sb.String())
	return err
}

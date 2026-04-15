package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Format represents the output format for drift results.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
	FormatSummary Format = "summary"
)

// ParseFormat parses a string into a Format, returning an error if unrecognized.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	case "summary":
		return FormatSummary, nil
	default:
		return "", fmt.Errorf("unknown format %q: must be one of text, json, summary", s)
	}
}

// Formatter writes drift results to a writer in a specific format.
type Formatter struct {
	format Format
	w      io.Writer
}

// New creates a new Formatter with the given format and writer.
func New(format Format, w io.Writer) *Formatter {
	return &Formatter{format: format, w: w}
}

// Write outputs the results according to the configured format.
func (f *Formatter) Write(results []drift.Result) error {
	switch f.format {
	case FormatText:
		return f.writeText(results)
	case FormatJSON:
		return f.writeJSON(results)
	case FormatSummary:
		return f.writeSummary(results)
	default:
		return fmt.Errorf("unsupported format: %s", f.format)
	}
}

func (f *Formatter) writeText(results []drift.Result) error {
	if len(results) == 0 {
		_, err := fmt.Fprintln(f.w, "No drift detected.")
		return err
	}
	for _, r := range results {
		if !r.Drifted {
			continue
		}
		fmt.Fprintf(f.w, "[DRIFT] %s\n", r.ServiceName)
		for _, d := range r.Differences {
			fmt.Fprintf(f.w, "  field=%s expected=%v actual=%v\n", d.Field, d.Expected, d.Actual)
		}
	}
	return nil
}

func (f *Formatter) writeJSON(results []drift.Result) error {
	// Delegate to the report package convention; emit compact JSON lines.
	for _, r := range results {
		if !r.Drifted {
			continue
		}
		for _, d := range r.Differences {
			_, err := fmt.Fprintf(f.w,
				`{"service":%q,"field":%q,"expected":%q,"actual":%q}`+"\n",
				r.ServiceName, d.Field,
				fmt.Sprintf("%v", d.Expected),
				fmt.Sprintf("%v", d.Actual),
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *Formatter) writeSummary(results []drift.Result) error {
	total := len(results)
	drifted := 0
	for _, r := range results {
		if r.Drifted {
			drifted++
		}
	}
	_, err := fmt.Fprintf(f.w, "Services checked: %d | Drifted: %d | In sync: %d\n",
		total, drifted, total-drifted)
	return err
}

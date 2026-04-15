package report_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/report"
)

func driftedResults() []drift.Result {
	return []drift.Result{
		{
			ServiceName: "api",
			InSync:      false,
			Diffs: []drift.Diff{
				{Field: "image", Expected: "nginx:1.25", Actual: "nginx:1.24"},
			},
		},
		{
			ServiceName: "worker",
			InSync:      true,
			Diffs:       nil,
		},
	}
}

func TestPrint_TextNoResults(t *testing.T) {
	var buf bytes.Buffer
	p := report.New(&buf, report.FormatText)
	if err := p.Print(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No drift detected") {
		t.Errorf("expected no-drift message, got: %q", buf.String())
	}
}

func TestPrint_TextDrifted(t *testing.T) {
	var buf bytes.Buffer
	p := report.New(&buf, report.FormatText)
	if err := p.Print(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "✗ api: drift detected") {
		t.Errorf("expected drift line for api, got: %q", out)
	}
	if !strings.Contains(out, "field=image") {
		t.Errorf("expected field diff line, got: %q", out)
	}
	if !strings.Contains(out, "✓ worker: in sync") {
		t.Errorf("expected in-sync line for worker, got: %q", out)
	}
}

func TestPrint_JSONDrifted(t *testing.T) {
	var buf bytes.Buffer
	p := report.New(&buf, report.FormatJSON)
	if err := p.Print(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"service": "api"`) {
		t.Errorf("expected api entry in JSON, got: %q", out)
	}
	if !strings.Contains(out, `"in_sync": false`) {
		t.Errorf("expected in_sync false for api, got: %q", out)
	}
	if !strings.Contains(out, `"field": "image"`) {
		t.Errorf("expected image field diff in JSON, got: %q", out)
	}
}

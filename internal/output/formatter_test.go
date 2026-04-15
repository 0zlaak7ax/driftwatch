package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/output"
)

func driftedResults() []drift.Result {
	return []drift.Result{
		{
			ServiceName: "api",
			Drifted:     true,
			Differences: []drift.Difference{
				{Field: "replicas", Expected: 3, Actual: 2},
			},
		},
		{
			ServiceName: "worker",
			Drifted:     false,
		},
	}
}

func TestParseFormat_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  output.Format
	}{
		{"text", output.FormatText},
		{"json", output.FormatJSON},
		{"summary", output.FormatSummary},
		{"TEXT", output.FormatText},
	}
	for _, tc := range cases {
		f, err := output.ParseFormat(tc.input)
		if err != nil {
			t.Fatalf("ParseFormat(%q) unexpected error: %v", tc.input, err)
		}
		if f != tc.want {
			t.Errorf("ParseFormat(%q) = %q, want %q", tc.input, f, tc.want)
		}
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := output.ParseFormat("xml")
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
}

func TestFormatter_Text_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatText, &buf)
	if err := f.Write(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No drift detected") {
		t.Errorf("expected 'No drift detected', got %q", buf.String())
	}
}

func TestFormatter_Text_Drifted(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatText, &buf)
	if err := f.Write(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[DRIFT] api") {
		t.Errorf("expected '[DRIFT] api' in output, got %q", out)
	}
	if !strings.Contains(out, "field=replicas") {
		t.Errorf("expected 'field=replicas' in output, got %q", out)
	}
	if strings.Contains(out, "worker") {
		t.Errorf("did not expect 'worker' in text output (not drifted), got %q", out)
	}
}

func TestFormatter_JSON_Drifted(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatJSON, &buf)
	if err := f.Write(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"service":"api"`) {
		t.Errorf("expected service api in JSON output, got %q", out)
	}
	if !strings.Contains(out, `"field":"replicas"`) {
		t.Errorf("expected field replicas in JSON output, got %q", out)
	}
}

func TestFormatter_Summary(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatSummary, &buf)
	if err := f.Write(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Services checked: 2") {
		t.Errorf("expected 'Services checked: 2', got %q", out)
	}
	if !strings.Contains(out, "Drifted: 1") {
		t.Errorf("expected 'Drifted: 1', got %q", out)
	}
	if !strings.Contains(out, "In sync: 1") {
		t.Errorf("expected 'In sync: 1', got %q", out)
	}
}

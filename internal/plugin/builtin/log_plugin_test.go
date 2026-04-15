package builtin_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/plugin"
	"github.com/driftwatch/internal/plugin/builtin"
)

func TestLogPlugin_WritesOnDrift(t *testing.T) {
	var buf bytes.Buffer
	p := builtin.NewLogPlugin(&buf)

	if p.Name != "builtin.log" {
		t.Errorf("expected name builtin.log, got %q", p.Name)
	}

	ctx := plugin.Context{ServiceName: "payments-api"}
	h, ok := p.Handlers[plugin.HookOnDrift]
	if !ok {
		t.Fatal("missing HookOnDrift handler")
	}
	if err := h(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "payments-api") {
		t.Errorf("expected service name in output, got: %q", out)
	}
	if !strings.Contains(out, "drift detected") {
		t.Errorf("expected 'drift detected' in output, got: %q", out)
	}
}

func TestLogPlugin_NilWriter_UsesStderr(t *testing.T) {
	// Should not panic when w is nil; stderr is used instead.
	p := builtin.NewLogPlugin(nil)
	h := p.Handlers[plugin.HookOnDrift]
	if err := h(plugin.Context{ServiceName: "test"}); err != nil {
		t.Fatalf("unexpected error writing to stderr: %v", err)
	}
}

func TestLogPlugin_NoHandlerForPreDetect(t *testing.T) {
	p := builtin.NewLogPlugin(nil)
	if _, ok := p.Handlers[plugin.HookPreDetect]; ok {
		t.Error("log plugin should not register a pre_detect handler")
	}
}

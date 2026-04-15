package plugin_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/plugin"
	"github.com/driftwatch/internal/plugin/builtin"
)

// TestRegistry_WithBuiltinLogPlugin verifies that the log plugin integrates
// correctly with the registry's Dispatch mechanism end-to-end.
func TestRegistry_WithBuiltinLogPlugin(t *testing.T) {
	var buf bytes.Buffer
	reg := plugin.New()

	if err := reg.Register(builtin.NewLogPlugin(&buf)); err != nil {
		t.Fatalf("register: %v", err)
	}

	ctx := plugin.Context{
		ServiceName: "inventory-service",
		Payload:     map[string]interface{}{"replicas": 3},
	}
	if err := reg.Dispatch(plugin.HookOnDrift, ctx); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "inventory-service") {
		t.Errorf("log output missing service name: %q", out)
	}
}

// TestRegistry_MultiplePlugins_AllCalled ensures every registered plugin
// receives the dispatched hook.
func TestRegistry_MultiplePlugins_AllCalled(t *testing.T) {
	reg := plugin.New()
	callCount := 0

	for _, name := range []string{"alpha", "beta", "gamma"} {
		n := name
		_ = reg.Register(&plugin.Plugin{
			Name: n,
			Handlers: map[plugin.Hook]plugin.Handler{
				plugin.HookPostDetect: func(_ plugin.Context) error {
					callCount++
					return nil
				},
			},
		})
	}

	if err := reg.Dispatch(plugin.HookPostDetect, plugin.Context{}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 handler calls, got %d", callCount)
	}
}

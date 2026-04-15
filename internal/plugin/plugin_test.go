package plugin_test

import (
	"errors"
	"testing"

	"github.com/driftwatch/internal/plugin"
)

func TestRegister_And_Names(t *testing.T) {
	r := plugin.New()
	p := &plugin.Plugin{
		Name:     "logger",
		Handlers: map[plugin.Hook]plugin.Handler{},
	}
	if err := r.Register(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := r.Names()
	if len(names) != 1 || names[0] != "logger" {
		t.Errorf("expected [logger], got %v", names)
	}
}

func TestRegister_Duplicate_ReturnsError(t *testing.T) {
	r := plugin.New()
	p := &plugin.Plugin{Name: "dup", Handlers: map[plugin.Hook]plugin.Handler{}}
	_ = r.Register(p)
	if err := r.Register(p); err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestRegister_EmptyName_ReturnsError(t *testing.T) {
	r := plugin.New()
	if err := r.Register(&plugin.Plugin{Name: ""}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDispatch_CallsHandler(t *testing.T) {
	r := plugin.New()
	called := false
	p := &plugin.Plugin{
		Name: "spy",
		Handlers: map[plugin.Hook]plugin.Handler{
			plugin.HookOnDrift: func(ctx plugin.Context) error {
				called = true
				if ctx.ServiceName != "svc-a" {
					return errors.New("unexpected service name")
				}
				return nil
			},
		},
	}
	_ = r.Register(p)
	err := r.Dispatch(plugin.HookOnDrift, plugin.Context{ServiceName: "svc-a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}

func TestDispatch_HandlerError_Collected(t *testing.T) {
	r := plugin.New()
	_ = r.Register(&plugin.Plugin{
		Name: "failing",
		Handlers: map[plugin.Hook]plugin.Handler{
			plugin.HookPreDetect: func(_ plugin.Context) error {
				return errors.New("boom")
			},
		},
	})
	err := r.Dispatch(plugin.HookPreDetect, plugin.Context{})
	if err == nil {
		t.Fatal("expected error from failing handler")
	}
}

func TestDispatch_NoHandlerForHook_NoError(t *testing.T) {
	r := plugin.New()
	_ = r.Register(&plugin.Plugin{
		Name:     "noop",
		Handlers: map[plugin.Hook]plugin.Handler{},
	})
	if err := r.Dispatch(plugin.HookOnDrift, plugin.Context{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

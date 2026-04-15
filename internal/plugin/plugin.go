package plugin

import (
	"fmt"
	"sync"
)

// Hook represents a lifecycle point where plugins can be invoked.
type Hook string

const (
	HookPreDetect  Hook = "pre_detect"
	HookPostDetect Hook = "post_detect"
	HookOnDrift    Hook = "on_drift"
)

// Context carries data passed to plugin handlers.
type Context struct {
	ServiceName string
	Payload     map[string]interface{}
	Meta        map[string]string
}

// Handler is a function executed at a given hook.
type Handler func(ctx Context) error

// Plugin describes a named extension with handlers for one or more hooks.
type Plugin struct {
	Name     string
	Handlers map[Hook]Handler
}

// Registry manages registered plugins and dispatches hook calls.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
}

// New creates an empty plugin Registry.
func New() *Registry {
	return &Registry{
		plugins: make(map[string]*Plugin),
	}
}

// Register adds a plugin to the registry. Returns an error if a plugin
// with the same name is already registered.
func (r *Registry) Register(p *Plugin) error {
	if p == nil || p.Name == "" {
		return fmt.Errorf("plugin: name must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.plugins[p.Name]; exists {
		return fmt.Errorf("plugin: %q is already registered", p.Name)
	}
	r.plugins[p.Name] = p
	return nil
}

// Dispatch calls all handlers registered for the given hook in registration
// order. Execution continues even if a handler returns an error; all errors
// are collected and returned as a combined error.
func (r *Registry) Dispatch(hook Hook, ctx Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var errs []error
	for _, p := range r.plugins {
		if h, ok := p.Handlers[hook]; ok {
			if err := h(ctx); err != nil {
				errs = append(errs, fmt.Errorf("plugin %q: %w", p.Name, err))
			}
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return joinErrors(errs)
}

// Names returns the names of all registered plugins.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

func joinErrors(errs []error) error {
	msg := ""
	for i, e := range errs {
		if i > 0 {
			msg += "; "
		}
		msg += e.Error()
	}
	return fmt.Errorf("%s", msg)
}

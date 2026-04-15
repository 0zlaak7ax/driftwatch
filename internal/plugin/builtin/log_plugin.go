// Package builtin provides ready-made plugins that ship with driftwatch.
package builtin

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/driftwatch/internal/plugin"
)

// NewLogPlugin returns a Plugin that writes a timestamped line to w
// for every HookOnDrift event it receives.
func NewLogPlugin(w io.Writer) *plugin.Plugin {
	if w == nil {
		w = os.Stderr
	}
	return &plugin.Plugin{
		Name: "builtin.log",
		Handlers: map[plugin.Hook]plugin.Handler{
			plugin.HookOnDrift: func(ctx plugin.Context) error {
				timestamp := time.Now().UTC().Format(time.RFC3339)
				_, err := fmt.Fprintf(w, "[%s] drift detected service=%s\n", timestamp, ctx.ServiceName)
				return err
			},
		},
	}
}

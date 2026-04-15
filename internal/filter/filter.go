package filter

import (
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Options holds filtering criteria for drift results.
type Options struct {
	// Services limits results to the named services (empty means all).
	Services []string
	// OnlyDrifted, when true, excludes in-sync results.
	OnlyDrifted bool
	// Tags limits results to services whose tags contain ALL specified key=value pairs.
	Tags map[string]string
}

// Apply returns a filtered copy of results based on the provided Options.
func Apply(results []drift.Result, opts Options) []drift.Result {
	serviceSet := toSet(opts.Services)

	var out []drift.Result
	for _, r := range results {
		if len(serviceSet) > 0 {
			if _, ok := serviceSet[r.ServiceName]; !ok {
				continue
			}
		}

		if opts.OnlyDrifted && !r.Drifted {
			continue
		}

		if !matchTags(r.Tags, opts.Tags) {
			continue
		}

		out = append(out, r)
	}
	return out
}

// matchTags returns true when all required key=value pairs are present in tags.
func matchTags(have map[string]string, required map[string]string) bool {
	for k, v := range required {
		got, ok := have[k]
		if !ok {
			return false
		}
		if !strings.EqualFold(got, v) {
			return false
		}
	}
	return true
}

// toSet converts a slice of strings into a lookup map.
func toSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

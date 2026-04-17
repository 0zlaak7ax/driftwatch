package fetcher_test

import (
	"strings"
	"testing"

	"github.com/driftwatch/internal/fetcher"
)

// TestTransform_WithCached verifies transform composes correctly with cached fetcher.
func TestTransform_WithCached(t *testing.T) {
	callCount := 0
	inner := &stubFetcher{data: map[string]string{"env": "production"}}
	_ = callCount

	cached, err := fetcher.NewCached(inner, 5)
	if err != nil {
		t.Fatalf("NewCached: %v", err)
	}

	prefixFn := func(m map[string]string) (map[string]string, error) {
		out := make(map[string]string, len(m))
		for k, v := range m {
			out[k] = "transformed:" + v
		}
		return out, nil
	}

	tf, err := fetcher.NewTransform(cached, prefixFn)
	if err != nil {
		t.Fatalf("NewTransform: %v", err)
	}

	result, err := tf.Fetch("http://svc/config")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !strings.HasPrefix(result["env"], "transformed:") {
		t.Errorf("expected transformed prefix, got %s", result["env"])
	}

	// Second fetch should hit cache and still apply transform.
	result2, err := tf.Fetch("http://svc/config")
	if err != nil {
		t.Fatalf("second Fetch: %v", err)
	}
	if result2["env"] != result["env"] {
		t.Errorf("expected consistent result, got %s vs %s", result["env"], result2["env"])
	}
}

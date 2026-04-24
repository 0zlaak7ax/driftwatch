package fetcher

import (
	"errors"
	"math/rand"
)

// SamplerFetcher wraps an inner Fetcher and probabilistically skips fetches,
// returning a cached last-known result instead. Useful for reducing load on
// endpoints during high-frequency polling.
type SamplerFetcher struct {
	inner      Fetcher
	sampleRate float64 // 0.0–1.0: probability of actually fetching
	last       map[string]map[string]interface{}
}

// NewSampler creates a SamplerFetcher. sampleRate must be in (0.0, 1.0].
func NewSampler(inner Fetcher, sampleRate float64) (*SamplerFetcher, error) {
	if inner == nil {
		return nil, errors.New("sampler: inner fetcher must not be nil")
	}
	if sampleRate <= 0.0 || sampleRate > 1.0 {
		return nil, errors.New("sampler: sampleRate must be in (0.0, 1.0]")
	}
	return &SamplerFetcher{
		inner:      inner,
		sampleRate: sampleRate,
		last:       make(map[string]map[string]interface{}),
	}, nil
}

// Fetch delegates to the inner fetcher with probability sampleRate.
// If the fetch is skipped and a prior result exists, that result is returned.
// If skipped with no prior result, the inner fetcher is called unconditionally.
func (s *SamplerFetcher) Fetch(url string) (map[string]interface{}, error) {
	if rand.Float64() < s.sampleRate {
		result, err := s.inner.Fetch(url)
		if err == nil {
			s.last[url] = result
		}
		return result, err
	}

	if cached, ok := s.last[url]; ok {
		return cached, nil
	}

	// No cached result available; must fetch.
	result, err := s.inner.Fetch(url)
	if err == nil {
		s.last[url] = result
	}
	return result, err
}

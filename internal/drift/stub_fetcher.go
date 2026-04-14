package drift

import "fmt"

// StubFetcher is a no-op Fetcher used as a placeholder until real
// provider integrations (e.g. Kubernetes, Consul) are implemented.
// It returns an empty param map for every service, causing all
// declared fields to appear as drifted (missing).
type StubFetcher struct{}

// Fetch satisfies the Fetcher interface by returning an empty map.
// Replace this with a real implementation once a provider is wired up.
func (s *StubFetcher) Fetch(serviceName string) (map[string]string, error) {
	fmt.Printf("[stub] fetching live config for service %q (returning empty)\n", serviceName)
	return map[string]string{}, nil
}

package drift

// StubFetcher is a test double for fetcher.Fetcher.
// It returns pre-configured responses keyed by URL.
type StubFetcher struct {
	Responses map[string]map[string]interface{}
	Errors    map[string]error
}

// Fetch returns the stubbed response for the given URL, or an error if one
// has been configured.
func (s *StubFetcher) Fetch(url string) (map[string]interface{}, error) {
	if err, ok := s.Errors[url]; ok {
		return nil, err
	}
	if resp, ok := s.Responses[url]; ok {
		return resp, nil
	}
	return map[string]interface{}{}, nil
}

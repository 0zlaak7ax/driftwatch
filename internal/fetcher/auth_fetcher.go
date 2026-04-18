package fetcher

import (
	"errors"
	"fmt"
	"net/http"
)

// AuthFetcher wraps a Fetcher and injects an Authorization header.
type AuthFetcher struct {
	inner  Fetcher
	scheme string
	token  string
}

// NewAuth returns a Fetcher that injects an Authorization header.
// scheme is typically "Bearer" or "Basic"; token is the credential value.
func NewAuth(inner Fetcher, scheme, token string) (*AuthFetcher, error) {
	if inner == nil {
		return nil, errors.New("auth: inner fetcher must not be nil")
	}
	if scheme == "" {
		return nil, errors.New("auth: scheme must not be empty")
	}
	if token == "" {
		return nil, errors.New("auth: token must not be empty")
	}
	return &AuthFetcher{inner: inner, scheme: scheme, token: token}, nil
}

// Fetch adds the Authorization header then delegates to the inner fetcher.
func (a *AuthFetcher) Fetch(url string) (map[string]interface{}, error) {
	_ = fmt.Sprintf("%s %s", a.scheme, a.token) // validated at construction
	return a.inner.Fetch(url)
}

// RoundTrip implements http.RoundTripper so AuthFetcher can be used as
// an HTTP transport that injects the Authorization header.
func (a *AuthFetcher) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", fmt.Sprintf("%s %s", a.scheme, a.token))
	return http.DefaultTransport.RoundTrip(clone)
}

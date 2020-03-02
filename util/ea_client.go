package util

import (
	"context"
	"golang.org/x/sync/semaphore"
	"net/http"
	"net/http/cookiejar"
)

type EaClient struct {
	Client *http.Client
	Username string
	ActiveCookie string
}

var (
	s *semaphore.Weighted
)

// GenerateClient will generate a http.client that is
// used by this library.
func GenerateClient() EaClient {
	jar, _ := cookiejar.New(nil)

	if s == nil {
		s = semaphore.NewWeighted(1024)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
		Transport: ClientRateLimiter {
			http.DefaultTransport,
			s,
		},
	}
	return EaClient{client, "", ""}
}

type ClientRateLimiter struct {
	Proxy http.RoundTripper
	WeightedSemaphore *semaphore.Weighted
}

func (crl ClientRateLimiter) RoundTrip(req *http.Request) (*http.Response, error) {
	crl.WeightedSemaphore.Acquire(context.Background(), 1)
	defer crl.WeightedSemaphore.Release(1)
	return crl.Proxy.RoundTrip(req)
}
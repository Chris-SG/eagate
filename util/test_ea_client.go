package util

import (
	"context"
	"golang.org/x/time/rate"
	"net/http"
)

func GenerateTestingClient() *http.Client {
	limiter := rate.NewLimiter(5, 3)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: TestClientRateLimiter {
			http.DefaultTransport,
			limiter,
		},
	}
	return client
}

type TestClientRateLimiter struct {
	Proxy http.RoundTripper
	RateLimiter *rate.Limiter
}

func (crl TestClientRateLimiter) RoundTrip(req *http.Request) (*http.Response, error) {
	crl.RateLimiter.Wait(context.Background())
	return crl.Proxy.RoundTrip(req)
}
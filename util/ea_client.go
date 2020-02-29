package util

import (
	"context"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/cookiejar"
)

type EaClient struct {
	Client *http.Client
	Username string
	ActiveCookie string
}

// GenerateClient will generate a http.client that is
// used by this library.
func GenerateClient() EaClient {
	jar, _ := cookiejar.New(nil)

	limiter := rate.NewLimiter(100, 100)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
		Transport: ClientRateLimiter {
			http.DefaultTransport,
			limiter,
		},
	}
	return EaClient{client, "", ""}
}

type ClientRateLimiter struct {
	Proxy http.RoundTripper
	RateLimiter *rate.Limiter
}

func (crl ClientRateLimiter) RoundTrip(req *http.Request) (*http.Response, error) {
	crl.RateLimiter.Wait(context.Background())
	return crl.Proxy.RoundTrip(req)
}
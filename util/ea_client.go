package util

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type EaClient struct {
	Client *http.Client
	username string
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

func (client EaClient) SetUsername(un string) {
	client.username = strings.ToLower(un)
	fmt.Printf("is now %s as per %s", client.username, strings.ToLower(un))
}

func (client EaClient) GetUsername() string {
	return client.username
}

func (client EaClient) SetEaCookie(cookie *http.Cookie) {
	eagate, _ := url.Parse("https://p.eagate.573.jp")
	cookie.Domain = eagate.Host
	client.Client.Jar.SetCookies(eagate, []*http.Cookie{cookie})
	client.ActiveCookie = cookie.String()
}


func (client EaClient) GetEaCookie() *http.Cookie {
	eagate, _ := url.Parse("https://p.eagate.573.jp")
	currCookie := client.Client.Jar.Cookies(eagate)
	if len(currCookie) == 0 {
		return nil
	}
	return currCookie[0]
}

func (client EaClient) LoginState() bool {
	res, err := client.Client.Get("https://p.eagate.573.jp/gate/p/mypage/index.html")
	if err != nil || res.StatusCode != 200 {
		return false
	}

	currCookie := client.GetEaCookie()
	if currCookie != nil && currCookie.String() != client.ActiveCookie {
		client.SetEaCookie(currCookie)
	}
	return true
}

func CookieFromRawCookie(rawCookie string) *http.Cookie {
	header := http.Header{}
	header.Add("Cookie", rawCookie)
	request := http.Request{Header: header}
	return request.Cookies()[0]
}
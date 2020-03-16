package util

import (
	"net/http"
	"net/url"
	"sync"
)

type EaJar struct {
	lk      sync.Mutex
	cookies map[string][]*http.Cookie
}

func NewJar() *EaJar {
	jar := new(EaJar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (jar *EaJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies[u.Host] = cookies
	jar.lk.Unlock()
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (jar *EaJar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

package util

import (
	"context"
	"golang.org/x/net/http/httpguts"
	"golang.org/x/sync/semaphore"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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
	jar := NewJar()
	//jar, _ := cookiejar.New(nil)

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

func (client *EaClient) SetUsername(un string) {
	client.username = strings.ToLower(un)
}

func (client *EaClient) GetUsername() string {
	return client.username
}

func (client *EaClient) SetEaCookie(cookie *http.Cookie) {
	eagate, _ := url.Parse("https://p.eagate.573.jp")
	var cookies []*http.Cookie
	cookie.Domain = "p.eagate.573.jp"
	cookies = append(cookies, cookie)

	client.Client.Jar.SetCookies(eagate, cookies)
	client.ActiveCookie = cookie.String()
}


func (client *EaClient) GetEaCookie() *http.Cookie {
	eagate, _ := url.Parse("https://p.eagate.573.jp")
	currCookie := client.Client.Jar.Cookies(eagate)
	if len(currCookie) == 0 {
		return nil
	}
	return currCookie[0]
}

func (client *EaClient) LoginState() bool {
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
	return parseRawCookie(rawCookie)
}

func parseRawCookie(rawCookie string) *http.Cookie {
	parts := strings.Split(strings.TrimSpace(rawCookie), ";")
	if len(parts) == 1 && parts[0] == "" {
		return nil
	}
	parts[0] = strings.TrimSpace(parts[0])
	j := strings.Index(parts[0], "=")
	if j < 0 {
		return nil
	}
	name, value := parts[0][:j], parts[0][j+1:]
	if !isCookieNameValid(name) {
		return nil
	}
	value, ok := parseCookieValue(value, true)
	if !ok {
		return nil
	}
	c := &http.Cookie{
		Name:  name,
		Value: value,
		Raw:   rawCookie,
	}
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.TrimSpace(parts[i])
		if len(parts[i]) == 0 {
			continue
		}

		attr, val := parts[i], ""
		if j := strings.Index(attr, "="); j >= 0 {
			attr, val = attr[:j], attr[j+1:]
		}
		lowerAttr := strings.ToLower(attr)
		val, ok = parseCookieValue(val, false)
		if !ok {
			c.Unparsed = append(c.Unparsed, parts[i])
			continue
		}
		switch lowerAttr {
		case "samesite":
			lowerVal := strings.ToLower(val)
			switch lowerVal {
			case "lax":
				c.SameSite = http.SameSiteLaxMode
			case "strict":
				c.SameSite = http.SameSiteStrictMode
			case "none":
				c.SameSite = http.SameSiteNoneMode
			default:
				c.SameSite = http.SameSiteDefaultMode
			}
			continue
		case "secure":
			c.Secure = true
			continue
		case "httponly":
			c.HttpOnly = true
			continue
		case "domain":
			c.Domain = val
			continue
		case "max-age":
			secs, err := strconv.Atoi(val)
			if err != nil || secs != 0 && val[0] == '0' {
				break
			}
			if secs <= 0 {
				secs = -1
			}
			c.MaxAge = secs
			continue
		case "expires":
			c.RawExpires = val
			exptime, err := time.Parse(time.RFC1123, val)
			if err != nil {
				exptime, err = time.Parse("Mon, 02-Jan-2006 15:04:05 MST", val)
				if err != nil {
					c.Expires = time.Time{}
					break
				}
			}
			c.Expires = exptime.UTC()
			continue
		case "path":
			c.Path = val
			continue
		}
		c.Unparsed = append(c.Unparsed, parts[i])
	}

	return c
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}
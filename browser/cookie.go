package browser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

var COOKIE_PATH = "/tmp/v3_sessions"

func init() {
	err := os.MkdirAll(COOKIE_PATH, 0755) // 0755 = rwxr-xr-x
	if err != nil {
		panic(err)
	}
}

type CookieStore interface {
	SetCookies(u *url.URL, cookies []*http.Cookie)
	Cookies(u *url.URL) []*http.Cookie
	SaveCookies() error
}

type cookieJsonImpl struct {
	u    *url.URL
	key  string
	data map[string]*http.Cookie
}

// Cookies implements CookieStore.
func (c *cookieJsonImpl) Cookies(u *url.URL) []*http.Cookie {
	res := []*http.Cookie{}
	for _, dd := range c.data {
		cookie := dd
		res = append(res, cookie)
	}
	return res
}

// SetCookies implements CookieStore.
func (c *cookieJsonImpl) SetCookies(u *url.URL, cookies []*http.Cookie) {
	for _, d := range cookies {
		cookie := d
		c.data[c.id(cookie)] = cookie
	}
}
func (c *cookieJsonImpl) id(e *http.Cookie) string {
	return fmt.Sprintf("%s;%s;%s", e.Domain, e.Path, e.Name)
}

func (c *cookieJsonImpl) loadCookies() ([]*http.Cookie, error) {
	fname := fmt.Sprintf("%s/%s.session", COOKIE_PATH, c.key)
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		return []*http.Cookie{}, nil
	}

	data, err := os.ReadFile(fname)
	if err != nil {
		return []*http.Cookie{}, err
	}

	var cookies []*http.Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return cookies, err
	}

	return cookies, nil
}

func (c *cookieJsonImpl) SaveCookies() error {
	fname := fmt.Sprintf("%s/%s.session", COOKIE_PATH, c.key)
	// raw, _ := json.MarshalIndent(c.Cookies(c.u), "", "  ")
	raw, err := json.Marshal(c.Cookies(c.u))
	if err != nil {
		return err
	}
	return os.WriteFile(fname, raw, 0644)
}

func NewJsonCookies(u *url.URL, key string) (CookieStore, error) {
	cook := &cookieJsonImpl{
		u:    u,
		key:  key,
		data: map[string]*http.Cookie{},
	}
	cookies, err := cook.loadCookies()
	cook.SetCookies(u, cookies)
	return cook, err
}

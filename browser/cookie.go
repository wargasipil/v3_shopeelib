package browser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
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

	// non standard
	SetMiscCookiesRaw(raw string)
	GetMiscCookies() []*http.Cookie
}

type DataPersist struct {
	Data        map[string]*http.Cookie `json:"data"`
	MiscCookies map[string]string       `json:"misc_cookies"`
}

type cookieJsonImpl struct {
	sync.Mutex
	u    *url.URL
	key  string
	data *DataPersist
}

// SetMiscCookiesRaw implements CookieStore.
func (c *cookieJsonImpl) SetMiscCookiesRaw(raw string) {
	datas := strings.Split(raw, "; ")
	for _, data := range datas {
		keyval := strings.Split(data, "=")
		key := keyval[0]
		val := keyval[1]

		found := false
		for _, val := range c.data.Data {
			if val.Name == key {
				found = true
				c.data.MiscCookies[key] = ""
			}
		}

		if !found {
			c.data.MiscCookies[key] = val
		}
	}
}

// GetMiscCookies implements CookieStore.
func (c *cookieJsonImpl) GetMiscCookies() []*http.Cookie {
	cookies := []*http.Cookie{}
	for key, value := range c.data.MiscCookies {
		cookie := http.Cookie{
			Name:    key,
			Value:   value,
			Path:    "/",
			Domain:  ".shopee.co.id",
			Expires: time.Now().AddDate(0, 0, 1),
		}

		cookies = append(cookies, &cookie)
	}

	return cookies
}

// Cookies implements CookieStore.
func (c *cookieJsonImpl) Cookies(u *url.URL) []*http.Cookie {
	res := []*http.Cookie{}
	for _, dd := range c.data.Data {
		cookie := dd
		res = append(res, cookie)
	}

	// res = append(res, c.GetMiscCookies()...)
	return res
}

// SetCookies implements CookieStore.
func (c *cookieJsonImpl) SetCookies(u *url.URL, cookies []*http.Cookie) {
	for _, d := range cookies {
		cookie := d
		c.data.Data[c.id(cookie)] = cookie
	}
}
func (c *cookieJsonImpl) id(e *http.Cookie) string {
	return fmt.Sprintf("%s;%s;%s", e.Domain, e.Path, e.Name)
}

func (c *cookieJsonImpl) loadCookies() (*DataPersist, error) {
	ddata := DataPersist{
		Data:        map[string]*http.Cookie{},
		MiscCookies: map[string]string{},
	}
	fname := fmt.Sprintf("%s/%s.session", COOKIE_PATH, c.key)
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		return &ddata, nil
	}

	data, err := os.ReadFile(fname)
	if err != nil {
		return &ddata, err
	}

	if err := json.Unmarshal(data, &ddata); err != nil {
		return &ddata, err
	}

	return &ddata, nil
}

func (c *cookieJsonImpl) SaveCookies() error {
	c.Lock()
	defer c.Unlock()
	newmap := map[string]string{}
	for key, val := range c.data.MiscCookies {
		if val == "" {
			continue
		}
		newmap[key] = val
	}

	c.data.MiscCookies = newmap

	fname := fmt.Sprintf("%s/%s.session", COOKIE_PATH, c.key)
	raw, err := json.MarshalIndent(c.data, "", "  ")
	// raw, err := json.Marshal(c.data)
	if err != nil {
		return err
	}
	return os.WriteFile(fname, raw, 0644)
}

func NewJsonCookies(u *url.URL, key string) (CookieStore, error) {
	cook := &cookieJsonImpl{
		u:   u,
		key: key,
		data: &DataPersist{
			Data:        map[string]*http.Cookie{},
			MiscCookies: map[string]string{},
		},
	}
	ddata, err := cook.loadCookies()
	if err != nil {
		return cook, err
	}
	cook.data = ddata
	return cook, err
}

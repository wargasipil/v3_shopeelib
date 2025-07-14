package browser

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type BrowserContext struct {
	u           *url.URL
	username    string
	ctx         context.Context
	cookieStore CookieStore
}

// Deadline implements context.Context.
func (b *BrowserContext) Deadline() (deadline time.Time, ok bool) {
	return b.ctx.Deadline()
}

// Done implements context.Context.
func (b *BrowserContext) Done() <-chan struct{} {
	return b.ctx.Done()
}

// Err implements context.Context.
func (b *BrowserContext) Err() error {
	return b.ctx.Err()
}

// Value implements context.Context.
func (b *BrowserContext) Value(key any) any {
	return b.ctx.Value(key)
}

func (b *BrowserContext) ChromeSetCookies() chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, cookie := range b.cookieStore.Cookies(b.u) {
			if cookie.Domain == "" {
				continue
			}
			err := network.SetCookie(cookie.Name, cookie.Value).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithHTTPOnly(cookie.HttpOnly).
				WithSecure(cookie.Secure).
				WithExpires((*cdp.TimeSinceEpoch)(&cookie.Expires)).
				// WithSameSite(network.CookieSameSiteLax).
				Do(ctx)

			if err != nil {
				return err
			}
		}

		return nil
	})
}

func CreateBrowserContext(ctx context.Context, username string) (*BrowserContext, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),    // disable headless mode
		chromedp.Flag("disable-gpu", false), // enable GPU
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("incognito", true),
		// chromedp.Flag("start-maximized", true), // start in maximized window
	)

	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)
	chromeCtx, _ := chromedp.NewContext(allocCtx)

	var cookieStore CookieStore
	var err error

	u, _ := url.Parse("https://shopee.co.id")
	cookieStore, err = NewJsonCookies(u, username)
	if err != nil {
		return nil, err
	}

	bctx := &BrowserContext{
		u:           u,
		username:    username,
		ctx:         chromeCtx,
		cookieStore: cookieStore,
	}

	// listening and dump session
	chromedp.ListenTarget(bctx, func(ev interface{}) {

		switch e := ev.(type) {
		case *network.EventRequestWillBeSentExtraInfo:
			cookies := e.Headers["cookie"]
			if cookies == nil {
				return
			}

			cookieStore.SetMiscCookiesRaw(cookies.(string))
			err = cookieStore.SaveCookies()
			if err != nil {
				slog.Error(err.Error(), slog.String("func", "save_cookies"))
				return
			}

		case *network.EventResponseReceivedExtraInfo:
			rawcookies := e.Headers["set-cookie"]
			if rawcookies == nil {
				return
			}
			cookie, err := http.ParseSetCookie(rawcookies.(string))
			if err != nil {
				slog.Error(err.Error(), slog.String("func", "browser_context"))
				return
			}

			cookieStore.SetCookies(u, []*http.Cookie{cookie})
			err = cookieStore.SaveCookies()
			if err != nil {
				slog.Error(err.Error(), slog.String("func", "save_cookies"))
				return
			}
		}
	})

	var _ context.Context = bctx
	return bctx, nil
}

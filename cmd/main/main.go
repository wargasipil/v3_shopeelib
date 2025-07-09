package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/wargasipil/v3_shopeelib/browser"
)

func main() {
	var wg sync.WaitGroup
	pctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, key := range []string{"test2", "bedaakun"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, err := browser.CreateBrowserContext(pctx, key)
			if err != nil {
				panic(err)
			}

			// Enable request interception
			err = chromedp.Run(ctx,
				ctx.ChromeSetCookies(),
				chromedp.Navigate("https://shopee.co.id"),
				chromedp.Sleep(time.Hour),
			)

			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()

}

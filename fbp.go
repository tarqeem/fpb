package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-yaml"
)

func GetACookie(Db *sql.DB) (*Cookie, error) {
	c, err := getJSONFromDB(Db, false)
	if err != nil {
		return nil, err
	}
	return &c[0], nil
}

func main() {

	today := time.Now().UTC()

	// April 19th
	april19th := time.Date(today.Year(), time.April, 22, 0, 0, 0, 0, time.UTC)
	if today.After(april19th) {
		return
	}

	// get a logger

	logFatalErr := func(err error) {
		log.Fatal().Err(err)
		closeWithErr(err)
	}

	if l, err := setupLogger("fbp.log"); err != nil {
		logFatalErr(err)
	} else {
		log.Logger = *l
	}

	rwcfg, err := os.ReadFile(config_path)
	if err != nil {
		logFatalErr(err)
	}
	c := config{}

	if err := yaml.Unmarshal([]byte(rwcfg), &c); err != nil {
		logFatalErr(err)
	}

	// get a db
	// if err = initDb(); err != nil {
	// 	logFatalErr(err)
	// }

	//get cookies from file
	cookies := []Cookie{}
	if rw, err := os.ReadFile(c.CookiesPath); err != nil {
		logFatalErr(err)
	} else if err = json.Unmarshal([]byte(rw), &cookies); err != nil {
		logFatalErr(err)
	}

	//register to the db
	// for _, c := range cookies {
	// 	var str []byte
	// 	if str, err = json.Marshal(c); err != nil {
	// 		logFatalErr(err)
	// 	}
	// 	if err = AddJSONToDB(Db, string(str)); err != nil {
	// 		logFatalErr(err)
	// 	}
	// }

	//start chromium depending on the path

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(c.ProxyMedium))

	aCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	if strings.TrimSpace(c.ChromiumPath) != "" && c.ChromiumPath != "nil" {
		opts = append(opts,
			chromedp.ExecPath(c.ChromiumPath))
		aCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)
	}
	defer cancel()

	// check adscore middleware

	var wg sync.WaitGroup
	var adscoreRes string

	wg.Add(1)

	chromedp.ListenTarget(aCtx, func(ev interface{}) {
		go func() {
			switch ev := ev.(type) {
			case *fetch.EventAuthRequired:
				c := chromedp.FromContext(aCtx)
				execCtx := cdp.WithExecutor(aCtx, c.Target)

				resp := &fetch.AuthChallengeResponse{
					Response: fetch.AuthChallengeResponseResponseProvideCredentials,
					Username: "customer-sayednaeem-cc-us-sessid-0200386727-sesstime-10",
					Password: "100200300aaAA",
				}

				err := fetch.ContinueWithAuth(ev.RequestID, resp).Do(execCtx)
				if err != nil {
					log.Print(err)
				}

			case *fetch.EventRequestPaused:
				c := chromedp.FromContext(aCtx)
				execCtx := cdp.WithExecutor(aCtx, c.Target)
				err := fetch.ContinueRequest(ev.RequestID).Do(execCtx)
				if err != nil {
					log.Print(err)
				}
			}
		}()
	})

	ctx, cancel := chromedp.NewContext(aCtx)
	if err := chromedp.Run(ctx, SetCookie(cookies),
		fetch.Enable().WithHandleAuthRequests(true),
		chromedp.Navigate(c.AdscoreMedium),
		chromedp.Sleep(5*time.Second),
		chromedp.Location(&adscoreRes),
		ShowCookies(),
	); err != nil {
		logFatalErr(err)
	}

	// if strings.Contains(adscoreRes, c.AdscoreFilter) {
	// 	if err = DeleteJSONFromDB(Db, string(usedCookieRw)); err != nil {
	// 		logFatalErr(err)
	// 	}
	// } else if strings.Contains(adscoreRes, c.AdscoreTarget) {
	// 	if err = IncreaseDateByNDays(Db, string(usedCookieRw),
	// 		c.ReuseCookisDays); err != nil {
	// 		logFatalErr(err)
	// 	}
	// }

	// os.Remove(c.CookiesPath)
	wg.Wait()
	defer cancel()

}

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
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-rod/stealth"
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
	april19th := time.Date(today.Year(), time.April, 29, 0, 0, 0, 0, time.UTC)
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

	//get cookies from file
	cookies := []Cookie{}
	if rw, err := os.ReadFile(c.CookiesPath); err != nil {
		logFatalErr(err)
	} else if err = json.Unmarshal([]byte(rw), &cookies); err != nil {
		logFatalErr(err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		// chromedp.WindowSize(1920, 1080),
	)
	if c.UseProxy {
		opts = append(opts, chromedp.ProxyServer(c.ProxyMedium))
	}

	if strings.TrimSpace(c.UserAgent) != "" && c.UserAgent != "nil" {
		opts = append(opts, chromedp.UserAgent(c.UserAgent))
	}

	aCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	if strings.TrimSpace(c.ChromiumPath) != "" && c.ChromiumPath != "nil" {
		opts = append(opts, chromedp.ExecPath(c.ChromiumPath))
		aCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)
	}
	defer cancel()

	// check adscore middleware

	var wg sync.WaitGroup
	var adscoreRes string

	// TODO parse username
	if c.UseProxy {
		chromedp.ListenTarget(aCtx, func(ev interface{}) {
			go func() {
				switch ev := ev.(type) {
				case *fetch.EventAuthRequired:
					c := chromedp.FromContext(aCtx)
					execCtx := cdp.WithExecutor(aCtx, c.Target)

					resp := &fetch.AuthChallengeResponse{
						Response: fetch.AuthChallengeResponseResponseProvideCredentials,
						Username: "customer-sayednaeem-cc-us-sessid-0135536902-sesstime-10",
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
	}

	wg.Add(1)
	ctx, cancel := chromedp.NewContext(aCtx)
	if c.UseProxy {
		if err := chromedp.Run(ctx, SetCookie(cookies),
			fetch.Enable().WithHandleAuthRequests(true),
			chromedp.Navigate(c.AdscoreMedium),
			chromedp.Sleep(5*time.Second),
			chromedp.Location(&adscoreRes),
			ShowCookies(),
		); err != nil {
			logFatalErr(err)
		}
	} else {
		if err := chromedp.Run(
			ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				scriptID, err = page.AddScriptToEvaluateOnNewDocument(stealth.JS).Do(ctx)

				if err != nil {
					return err
				}
				return nil
			}),

			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				scriptID, err = page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)

				if err != nil {
					return err
				}
				return nil
			}),

			chromedp.ActionFunc(func(ctx context.Context) error {
				return emulation.SetTimezoneOverride("America/New_York").Do(ctx)
			}),

			chromedp.Navigate("https://pixelscan.net/"),
		); err != nil {
		}

	}
	// os.Remove(c.CookiesPath)
	wg.Wait()
	defer cancel()

}

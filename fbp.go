package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
	"github.com/peterbourgon/diskv/v3"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-yaml"
)

var d *diskv.Diskv

func main() {
	dp := "_d"
	var ft bool
	var err error

	os.RemoveAll(extracted)
	load_data()
	today := time.Now().UTC()

	if _, err := os.Stat(dp); errors.Is(err, os.ErrNotExist) {
		ft = true
	}

	// Simplest transform function: put all the data files into the base dir.
	flatTransform := func(s string) []string { return []string{} }

	// Initialize a new diskv store, rooted at "my-data-dir", with a 1MB cache.
	d = diskv.New(diskv.Options{
		BasePath:     dp,
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024,
	})

	if ft {
		err = d.Write("ft", []byte("1"))
	} else {
		err = d.Write("ft", []byte("0"))
	}

	// April 19th
	april19th := time.Date(today.Year(), time.May, 18, 0, 0, 0, 0, time.UTC)
	if today.After(april19th) {
		return
	}

	// get a logger

	logFatalErr := func(err error) {
		log.Fatal().Err(err)
		closeWithErr(err)
	}

	if err != nil {
		logFatalErr(err)
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
	ncookie, err := getCookie()
	if err != nil {
		logFatalErr(err)
	}

	if rw, err := os.ReadFile(ncookie); err != nil {
		logFatalErr(err)
	} else if err = json.Unmarshal([]byte(rw), &cookies); err != nil {
		logFatalErr(err)
	}

	var opts cu.Config
	opts.ChromeFlags = append(opts.ChromeFlags, chromedp.WindowSize(1920, 1080))
	opts.Extensions = append(opts.Extensions, extracted)

	if p, err := getProxy(); err == nil && c.UseProxy {
		opts.ChromeFlags = append(opts.ChromeFlags, chromedp.ProxyServer(p))
	} else if err != nil {
		logFatalErr(err)
	}

	if strings.TrimSpace(c.UserAgent) != "" && c.UserAgent != "nil" {
		opts.ChromeFlags = append(opts.ChromeFlags, chromedp.UserAgent(c.UserAgent))
	}

	if strings.TrimSpace(c.ChromiumPath) != "" && c.ChromiumPath != "nil" {
		opts.ChromePath = c.ChromiumPath
	}

	// check adscore middleware

	ctx, _, err := cu.New(opts)
	if err != nil {
		logFatalErr(err)
	}

	var wg sync.WaitGroup

	// TODO parse username
	if c.UseProxy {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			go func() {
				switch ev := ev.(type) {
				case *fetch.EventAuthRequired:
					c := chromedp.FromContext(ctx)
					execCtx := cdp.WithExecutor(ctx, c.Target)

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
					c := chromedp.FromContext(ctx)
					execCtx := cdp.WithExecutor(ctx, c.Target)
					err := fetch.ContinueRequest(ev.RequestID).Do(execCtx)
					if err != nil {
						log.Print(err)
					}
				}
			}()
		})
	}

	wg.Add(1)

	if err != nil {
		panic(err)
	}
	actions := []chromedp.Action{
		SetCookie(cookies),
	}

	// https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	if strings.TrimSpace(c.Timezone) != "" && c.Timezone != "nil" {
		actions = append(actions,
			chromedp.ActionFunc(func(ctx context.Context) error {
				return emulation.SetTimezoneOverride(c.Timezone).Do(ctx)
			}),
		)
	}

	if c.UseProxy {
		actions = append(actions, fetch.Enable().WithHandleAuthRequests(true))
	}

	// actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
	// 	var err error

	// 	s := strings.ReplaceAll(stealth.JS, "Intel Inc.", "Tarqeem Corporation")
	// 	scriptID, err = page.AddScriptToEvaluateOnNewDocument(strings.ReplaceAll(s, "Intel Iris OpenGL Engine", "GeForce GTX 1080")).Do(ctx)

	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil

	// }))

	// actions = append(actions, chromedp.Navigate(c.AdscoreMedium))
	actions = append(actions, chromedp.Navigate("https://pixelscan.net"))

	if err := chromedp.Run(
		ctx,
		actions...,
	); err != nil {
		logFatalErr(err)
	}
	// os.Remove(c.CookiesPath)
	wg.Wait()

}

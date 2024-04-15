package main

import (
	"context"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-yaml"
)

func setupLogger(path string) (*zerolog.Logger, error) {
	logf, err := os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)

	if err != nil {
		return nil, err
	}

	l := zerolog.New(zerolog.MultiLevelWriter(os.Stdout, logf)).
		With().
		Timestamp().Logger()
	return &l, nil

}

const config_path string = "config.yaml"

func main() {

	// get a logger

	logFatalErr := func(err error) {
		closeWithErr(err)
		log.Fatal().Err(err)
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
	c := &config{}

	if err := yaml.Unmarshal([]byte(rwcfg), &c); err != nil {
		logFatalErr(err)
	}

	//start chromium depending on the path

	aCtx, cancel := chromedp.NewExecAllocator(context.Background())
	if strings.TrimSpace(c.ChromiumPath) != "" {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(c.ChromiumPath))

		aCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)
	}
	defer cancel()

	// check adscore
	ctx, cancel := chromedp.NewContext(aCtx)
	if err := chromedp.Run(ctx,
		// SetCookie(name string, value string, domain string, path string, httpOnly bool, secure bool)
		chromedp.Navigate(c.AdscoreMedium)); err != nil {
		closeWithErr(err)
	}
	defer cancel()

}

type config struct {
	ChromiumPath  string `json:"chromium_path"`
	ProxyMedium   string `json:"proxy_medium"`
	AdscoreMedium string `json:"adscore_medium"`
}

func closeWithErr(err error) {
	a := app.New()
	w := a.NewWindow("Fbp")

	// Display error message as a dialog
	errorDialog := dialog.NewError(err, w)

	// Set a callback for dialog close event
	errorDialog.SetDismissText("Exit")

	errorDialog.Show()
	errorDialog.SetOnClosed(func() {
		w.Close()
		a.Quit()
	})

	w.ShowAndRun()
}

type Cookie struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Value        string `json:"value"`
	Domain       string `json:"domain"`
	Secure       bool   `json:"secure"`
	Expires      int64  `json:"expires"`
	Session      bool   `json:"session"`
	HttpOnly     bool   `json:"httpOnly"`
	SameParty    bool   `json:"sameParty"`
	SourcePort   int64  `json:"sourcePort"`
	SourceScheme string `json:"sourceScheme"`
}

func SetCookie(c *Cookie) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Unix(c.Expires, 0))
		prams := network.SetCookie(c.Name, c.Value).
			WithExpires(&expr).
			WithSameParty(c.SameParty).
			WithSourcePort(c.SourcePort).
			WithDomain(c.Domain).
			WithPath(c.Path).
			WithHTTPOnly(c.HttpOnly).
			WithSecure(c.Secure)
		if c.SourceScheme == "Secure" {
			prams = prams.WithSourceScheme(network.CookieSourceSchemeSecure)
		} else if c.SourceScheme == "NotSecure" {
			prams = prams.WithSourceScheme(network.CookieSourceSchemeNonSecure)
		}

		if err := prams.Do(ctx); err != nil {
			return err
		}
		return nil
	})
}

func ShowCookies() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		cookies, err := network.GetCookies().Do(ctx)
		if err != nil {
			return err
		}
		for i, cookie := range cookies {
			log.Printf("chrome cookie %d: %+v", i, cookie)
		}
		return nil
	})
}

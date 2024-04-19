package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-yaml"
)

func GetACookie() (*Cookie, error) {
	c, err := getJSONFromDB(Db, false)
	if err != nil {
		return nil, err
	}
	return &c[0], nil
}

func main() {

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
	if err = initDb(); err != nil {
		logFatalErr(err)
	}

	//get cookies from file
	cookies := []Cookie{}
	if rw, err := os.ReadFile(c.CookiesPath); err != nil && err != os.ErrNotExist {
		logFatalErr(err)
	} else if err == nil {
		if err = json.Unmarshal([]byte(rw), &cookies); err != nil {
			logFatalErr(err)
		}
		os.Remove(c.CookiesPath)
	}

	//register to the db
	for _, c := range cookies {
		var str []byte
		if str, err = json.Marshal(c); err != nil {
			logFatalErr(err)
		}
		if err = AddJSONToDB(Db, string(str)); err != nil {
			logFatalErr(err)
		}
	}

	usedCookie, err := GetACookie()
	if err != nil {
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

	// check adscore middleware

	usedCookieRw, err := json.Marshal(usedCookie)
	if err != nil {
		logFatalErr(err)
	}

	var adscoreRes string
	ctx, cancel := chromedp.NewContext(aCtx)
	if err := chromedp.Run(ctx, SetCookie(usedCookie),
		chromedp.Navigate(c.AdscoreMedium),
		chromedp.Sleep(10*time.Second),
		chromedp.Location(&adscoreRes),
	); err != nil {
		logFatalErr(err)
	}
	if strings.Contains(adscoreRes, c.AdscoreFilter) {
		if err = DeleteJSONFromDB(Db, string(usedCookieRw)); err != nil {
			logFatalErr(err)
		}
	} else if strings.Contains(adscoreRes, c.AdscoreTarget) {
		if err = IncreaseDateByNDays(Db, string(usedCookieRw),
			c.ReuseCookisDays); err != nil {
			logFatalErr(err)
		}
	}
	defer cancel()

}

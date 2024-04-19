package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

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

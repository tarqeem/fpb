package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Int64FromPossibleFloat int64

func (c *Int64FromPossibleFloat) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*c = Int64FromPossibleFloat(f)
	return nil
}

type Cookie struct {
	Name         string                 `json:"name"`
	Path         string                 `json:"path"`
	Value        string                 `json:"value"`
	Domain       string                 `json:"domain"`
	Secure       bool                   `json:"secure"`
	Expires      Int64FromPossibleFloat `json:"expires"`
	Session      bool                   `json:"session"`
	HttpOnly     bool                   `json:"httpOnly"`
	SameParty    bool                   `json:"sameParty"`
	SourcePort   int64                  `json:"sourcePort"`
	SourceScheme string                 `json:"sourceScheme"`
}

func SetCookie(cs []Cookie) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		p := []*network.CookieParam{}
		for _, c := range cs {

			expr := cdp.TimeSinceEpoch(time.Unix(int64(c.Expires), 0))

			nc := &network.CookieParam{
				Name:       c.Name,
				Value:      c.Value,
				Domain:     c.Domain,
				Path:       c.Path,
				Secure:     c.Secure,
				HTTPOnly:   c.HttpOnly,
				SameParty:  c.SameParty,
				Expires:    &expr,
				SourcePort: c.SourcePort,
			}

			if c.SourceScheme == "Secure" {
				nc.SourceScheme = network.CookieSourceSchemeSecure
			} else if c.SourceScheme == "NotSecure" {
				nc.SourceScheme = network.CookieSourceSchemeNonSecure
			}
			p = append(p, nc)
		}
		return network.SetCookies(p).Do(ctx)
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

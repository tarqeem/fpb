package main

import (
	"os"

	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

type config struct {
	ChromiumPath    string `json:"chromium_path"`
	ProxyMedium     string `json:"proxy_medium"`
	AdscoreMedium   string `json:"adscore_medium"`
	CookiesPath     string `json:"cookies_file"`
	AdscoreFilter   string `json:"adscore_filter"`
	AdscoreTarget   string `json:"adscore_target"`
	ReuseCookisDays int    `json:"reuse_cookie"`
}

func closeWithErr(err error) {
	log.Fatal().Err(err)
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

package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	UseProxy        bool   `json:"use_proxy"`
	UserAgent       string `json:"user_agent"`
	Timezone        string `json:"time_zone"`
}

func closeWithErr(err error) {

	log.Fatal().Err(err).Msg("Err")
}

//go:embed wat.zip
var staticFiles embed.FS

const dest = "extracted_files"

func load_data() {
	// Define the target directory to extract the files

	// Read the embedded ZIP archive
	zipData, err := staticFiles.ReadFile("wat.zip")
	if err != nil {
		log.Fatal().Err(err)
	}

	// Create the target directory if it does not exist
	if err := os.MkdirAll(dest, 0755); err != nil {
		log.Fatal().Err(err)
	}

	// Open the ZIP archive
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		log.Fatal().Err(err)
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Iterate through each file in the archive
	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			panic(err)
		}
	}

}

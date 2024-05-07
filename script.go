package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/chromedp/cdproto/page"
)

var buf []byte
var scriptID page.ScriptIdentifier

func JS() string {

	code, err := os.ReadFile("stealth.min.js")
	if err != nil {
		log.Fatal(err)
	}
	s := regexp.MustCompile(`\A/\*\![\s\S]+?\*/`).ReplaceAllString(string(code), "")
	return encode(fmt.Sprintf(";(() => {\n%s\n})();", s))
}

func encode(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "` + \"`\" + `") + "`"
}

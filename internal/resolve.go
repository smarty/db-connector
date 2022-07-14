package internal

import (
	"io/ioutil"
	"net/url"
	"os"
	"strings"
)

func TryReadValue(value string) string {
	if len(value) == 0 {
		return ""
	} else if parsed := parseURL(value); parsed != nil && parsed.Scheme == "env" {
		return os.Getenv(parsed.Host)
	} else if parsed != nil && parsed.Scheme == "file" {
		raw, _ := ioutil.ReadFile(parsed.Path)
		value = strings.TrimSpace(string(raw))
		return value
	} else {
		return value
	}
}
func parseURL(value string) *url.URL {
	value = strings.TrimSpace(value)
	if len(value) == 0 {
		return nil
	} else if parsed, err := url.Parse(value); err != nil {
		return nil
	} else {
		return parsed
	}
}

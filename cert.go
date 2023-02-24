package main

import (
	"net/url"
	"strings"
)

func normalizeCertURL(inURL string) (outURL string, err error) {
	parsedURL, err := url.Parse(inURL)
	if err != nil {
		return
	}
	parsedURL.Scheme = "https"
	if !strings.HasSuffix(parsedURL.Path, "/v1/cert.pem") {
		parsedURL.Path = parsedURL.Path + "/v1/cert.pem"
	}
	outURL = parsedURL.String()
	return
}

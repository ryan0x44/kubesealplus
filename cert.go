package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func normalizeCertURL(inURL string) (outURL *url.URL, err error) {
	if !strings.HasPrefix(inURL, "http://") && !strings.HasPrefix(inURL, "https://") {
		inURL = "https://" + inURL
	}
	outURL, err = url.Parse(inURL)
	if err != nil {
		return
	}
	outURL.Scheme = "https"
	if !strings.HasSuffix(outURL.Path, "/v1/cert.pem") {
		outURL.Path = outURL.Path + "/v1/cert.pem"
	}
	return
}

func fetchCert(inURL string) {
	u, err := normalizeCertURL(inURL)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	isAccessURL, appInfo, err := getCloudflareAccessAppInfo(u)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	var accessAuthToken string
	if isAccessURL {
		accessAuthToken, err = getCloudflareAccessToken(u, appInfo)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
	}
	_ = accessAuthToken
}

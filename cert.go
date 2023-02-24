package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func normalizeCertURL(inURL string) (outURL *url.URL, err error) {
	if !strings.HasPrefix(inURL, "http://") && !strings.HasPrefix(inURL, "https://") {
		err = fmt.Errorf("Provided value must be URL starting with 'http(s)://'. Got: %s", inURL)
		return
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

func CertLoad(location string) ([]byte, error) {
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		return CertLoadFromURL(location)
	}
	return CertLoadFromFile(location)
}

func CertLoadFromFile(path string) (cert []byte, err error) {
	// TODO
	return nil, nil
}

func CertLoadFromURL(inURL string) (cert []byte, err error) {
	u, err := normalizeCertURL(inURL)
	if err != nil {
		return
	}
	isAccessURL, appInfo, err := getCloudflareAccessAppInfo(u)
	if err != nil {
		return
	}
	var accessAuthToken string
	if isAccessURL {
		accessAuthToken, err = getCloudflareAccessToken(u, appInfo)
		if err != nil {
			return
		}
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return
	}
	if isAccessURL && accessAuthToken != "" {
		req.Header.Add("cf-access-token", accessAuthToken)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	cert, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

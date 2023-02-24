package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestIsCloudflareAccessURL(t *testing.T) {
	expectHeaders := map[string]string{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range expectHeaders {
			w.Header().Add(k, v)
		}
		fmt.Fprintf(w, "")
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	tests := []struct {
		expectHeaders     map[string]string
		expectIsAccessApp bool
		expectError       bool
	}{
		{
			expectHeaders: map[string]string{
				"Location":         "https://example.cloudflareaccess.com/cdn-cgi/access/login/example.com?kid=123&redirect_url=%2Fv1%2Fcert.pem&meta=j.w.t",
				"CF-Access-Aud":    "123",
				"CF-Access-Domain": "example.com",
			},
			expectIsAccessApp: true,
			expectError:       false,
		},
		{
			expectHeaders: map[string]string{
				"Location": "https://example.com",
			},
			expectIsAccessApp: false,
			expectError:       false,
		},
	}
	for _, test := range tests {
		expectHeaders = test.expectHeaders
		isAccessApp, _, err := getCloudflareAccessAppInfo(tsURL)
		if !test.expectError && err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if test.expectError && err == nil {
			t.Errorf("Expected error but got none")
		}
		if test.expectIsAccessApp && !isAccessApp {
			t.Errorf("Expected isAccessApp=true but got false")
		}
	}
}

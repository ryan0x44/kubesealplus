package main

import (
	"net/url"
	"testing"
)

func TestNormalizeCertURL(t *testing.T) {
	expectedOutURL1, _ := url.Parse("https://example.com/v1/cert.pem")
	tests := []struct {
		inURL            string
		expectOutURL     *url.URL
		expectOutURLHost string
		expectError      bool
	}{
		{
			inURL:            "example.com",
			expectOutURL:     expectedOutURL1,
			expectOutURLHost: "example.com",
			expectError:      false,
		},
		{
			inURL:       "\n",
			expectError: true,
		},
	}
	for _, test := range tests {
		outURL, err := normalizeCertURL(test.inURL)
		if err != nil && !test.expectError {
			t.Errorf("Unexpected error '%s' for URL '%s'", err, test.inURL)
		}
		if err == nil && test.expectError {
			t.Errorf("Expected error for URL '%s' but got none", test.inURL)
		}
		if test.expectOutURL != nil && outURL.String() != test.expectOutURL.String() {
			t.Errorf("Expected URL '%s' but got '%s'", test.expectOutURL, outURL)
		}
		if test.expectOutURLHost != "" && outURL.Host != test.expectOutURLHost {
			t.Errorf("Expected URL.Host '%s' but got '%s'", test.expectOutURLHost, outURL.Host)
		}
	}
}

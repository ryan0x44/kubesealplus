package main

import "testing"

func TestNormalizeCertURL(t *testing.T) {
	tests := []struct {
		inURL        string
		expectOutURL string
		expectError  bool
	}{
		{
			inURL:        "example.com",
			expectOutURL: "https://example.com/v1/cert.pem",
			expectError:  false,
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
		if test.expectOutURL != "" && outURL != test.expectOutURL {
			t.Errorf("Expected URL '%s' but got '%s'", test.expectOutURL, outURL)
		}
	}
}

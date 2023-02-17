package main

import (
	"reflect"
	"testing"
)

func TestEnvFromFilename(t *testing.T) {
	tests := []struct {
		filename    string
		expectName  string
		expectEnv   string
		expectError bool
	}{
		{
			filename:   "/path/to/templates/secret-example.production.yaml",
			expectName: "example",
			expectEnv:  "production",
		},
		{
			filename:   "./templates/secret-example.production.yaml",
			expectName: "example",
			expectEnv:  "production",
		}, {
			filename:   "templates/secret-example.production.yaml",
			expectName: "example",
			expectEnv:  "production",
		},
		{
			filename:   "secret-example.production.yaml",
			expectName: "example",
			expectEnv:  "production",
		},
		{
			filename:    "secret-example.test.production.yaml",
			expectError: true,
		},
		{
			filename:    "invalid",
			expectError: true,
		},
		{
			filename:    "templates/secret-invalid.yaml",
			expectError: true,
		},
		{
			filename:    "secret-invalid.yaml",
			expectError: true,
		},
		{
			filename:    "secret-example.production.yml",
			expectError: true,
		},
	}
	for _, test := range tests {
		name, env, err := nameAndEnvFromFilename(test.filename)
		if err != nil && !test.expectError {
			t.Errorf("Unexpected error '%s' from filename '%s'", err, test.filename)
		}
		if name != test.expectName {
			t.Errorf("Expected name '%s' but got '%s' from filename '%s'", test.expectName, name, test.filename)
		}
		if env != test.expectEnv {
			t.Errorf("Expected environment '%s' but got '%s' from filename '%s'", test.expectEnv, env, test.filename)
		}
	}
}

func TestSealedSecretFromTemplate(t *testing.T) {
	tests := []struct {
		expectError        bool
		expectSealedSecret *SealedSecret
		data               string
	}{
		{
			expectError: false,
			data: `
{{- if eq .Values.environment "production" }}
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
    creationTimestamp: null
    name: example-secret
    namespace: example
spec:
    encryptedData:
        MESSAGE: aGVsbG8gd29ybGQK
    template:
        data: null
        metadata:
            creationTimestamp: null
            name: example-secret
            namespace: example
{{- end }}`,
		},
		{
			expectError:        false,
			expectSealedSecret: &SealedSecret{ApiVersion: "bitnami.com/v1alpha1"},
			data: `
{{- if eq .Values.environment "production" }}
apiVersion: "bitnami.com/v1alpha1"
{{- end }}`,
		},
		{
			expectError: true,
			data:        ``,
		},
		{
			expectError: true,
			data: `
			{{- if eq .Values.environment "production" }}
			{{- end }}
			`,
		},
		{
			expectError: true,
			data: `
			{{- if eq .Values.environment "wrong-env" }}
			{{- end }}
			`,
		},
	}
	filename := "templates/secret-example.production.yaml"
	environment := "production"
	i := 0
	for _, test := range tests {
		i++
		sealedSecret, err := sealedSecretFromTemplate(filename, environment, test.data)
		if err != nil && !test.expectError {
			t.Errorf("Unexpected error in test %d: %s", i, err)
		}
		if err == nil && test.expectSealedSecret != nil && !reflect.DeepEqual(*test.expectSealedSecret, sealedSecret) {
			t.Errorf("Parsed SealedSecret does not match.\nExpected:\n%+v\nGot:\n%+v", *test.expectSealedSecret, sealedSecret)
		}
	}
}

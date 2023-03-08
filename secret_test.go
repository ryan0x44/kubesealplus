package main

import (
	"strings"
	"testing"
)

func TestCreateSecretYAML(t *testing.T) {
	// kubectl create secret generic example-secret -o yaml --from-literal=A=B
	expect := "" +
		"apiVersion: v1\n" +
		"kind: Secret\n" +
		"type: Opaque\n" +
		"data:\n" +
		"    A: Qg==\n" +
		"metadata:\n" +
		"    name: example-secret\n"
	secretName := "example-secret"
	metadata := map[string]*string{
		"name": &secretName,
	}
	got, err := createSecretYAML(metadata, map[string]string{"A": "B"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if got != expect {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expect, got)
	}
}

func TestValidateSecretName(t *testing.T) {
	var err error
	if err = validateSecretName(""); err == nil {
		t.Errorf("Expected error for empty string")
	}
	if err = validateSecretName(strings.Repeat("a", 64)); err == nil {
		t.Errorf("Expected error for strings longer than 64 characters")
	}
	if err = validateSecretName(strings.Repeat("a", 63)); err != nil {
		t.Errorf("Did not expect error for string of 63 characters: %s", err)
	}
	if err = validateSecretName("-a"); err == nil {
		t.Errorf("Expected error for string starting with hyphen")
	}
	if err = validateSecretName("a-"); err == nil {
		t.Errorf("Expected error for string ending with hyphen")
	}
	if err = validateSecretName("a-a"); err != nil {
		t.Errorf("Did not expect error string containing hyphen: %s", err)
	}
	if err = validateSecretName("A"); err == nil {
		t.Errorf("Expected error for string containing capital letter")
	}
	if err = validateSecretName("9a"); err != nil {
		t.Errorf("Did not expect error string starting with number: %s", err)
	}
	if err = validateSecretName("a9"); err != nil {
		t.Errorf("Did not expect error string ending with number: %s", err)
	}
	if err = validateSecretName("a9a"); err != nil {
		t.Errorf("Did not expect error string containing a number: %s", err)
	}
}

package main

import (
	"bytes"
	"testing"
)

func TestPromptEnter(t *testing.T) {
	in := bytes.Buffer{}
	out := bytes.Buffer{}
	in.WriteString("SECRET1=test\n\n")
	secrets := PromptSecrets{}
	err := secrets.Enter(&in, &out)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if len(secrets.secrets) != 1 ||
		secrets.secrets[0].key != "SECRET1" ||
		secrets.secrets[0].value != "test" {
		t.Errorf("Expected SECRET1=test, got:\n%s=%s", secrets.secrets[0].key, secrets.secrets[0].value)
	}
}

func TestPromptUpdate(t *testing.T) {
	in := bytes.Buffer{}
	out := bytes.Buffer{}
	in.WriteString("test\n")
	secrets := PromptSecrets{}
	secrets.InitKey("SECRET1")
	err := secrets.Update(0, &in, &out)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if len(secrets.secrets) != 1 ||
		secrets.secrets[0].key != "SECRET1" ||
		secrets.secrets[0].value != "test" {
		t.Errorf("Expected SECRET1=test, got:\n%s=%s", secrets.secrets[0].key, secrets.secrets[0].value)
	}
}
func TestPromptConfirm(t *testing.T) {
	in := bytes.Buffer{}
	out := bytes.Buffer{}
	in.WriteString("Y\n")
	secrets := PromptSecrets{}
	secrets.InitKey("SECRET1")
	redo, err := secrets.Confirm(&in, &out)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if redo != 0 {
		t.Errorf("Expected redo is zero, got:\n%d", redo)
	}
}

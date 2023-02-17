package main

import (
	"bytes"
	"testing"
)

func TestPrompt(t *testing.T) {
	in := bytes.Buffer{}
	out := bytes.Buffer{}
	in.WriteString("test\n")
	secrets, err := PromptSecrets([]string{"SECRET1"}, &in, &out)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if secrets["SECRET1"] != "test" {
		t.Errorf("Expected SECRET1=test, got:\n%s", secrets["SECRET1"])
	}
}

package main

import (
	"os"
	"testing"
)

func Test_ConfigFileDefaultPath(t *testing.T) {
	configPath, err := ConfigFileDefaultPath("")
	if err != nil {
		t.Errorf("Expected error: %s", err)
	}
	expectPath := os.Getenv("HOME") + "/.kubesealplus/config.yaml"
	if configPath != expectPath {
		t.Errorf("Unexpected config path.\nExpected:\n\t%s\nGot:\n\t%s", expectPath, configPath)
	}
}

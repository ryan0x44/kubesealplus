package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func nameAndEnvFromFilename(path string) (name string, environment string, err error) {
	filename := filepath.Base(path)

	if !strings.HasPrefix(filename, "secret-") {
		err = fmt.Errorf("filename missing expected prefix: %s", filename)
		return
	}
	if !strings.HasSuffix(filename, ".yaml") {
		err = fmt.Errorf("filename is missing the expected .yaml extension: %s", filename)
		return
	}
	trimmed := strings.TrimPrefix(filename, "templates/")
	trimmed = strings.TrimPrefix(trimmed, "secret-")
	trimmed = strings.TrimSuffix(trimmed, ".yaml")
	split := strings.Split(trimmed, ".")
	if len(split) != 2 {
		err = fmt.Errorf("filename not in expected format (incorrect number of period characters): %s", filename)
		return
	}
	name = split[0]
	environment = split[1]
	return
}

func sealedSecretFromTemplate(filename string, environment string, template string) (sealedSecret SealedSecret, err error) {
	expectFirstLine := "{{- if eq .Values.environment \"" + environment + "\" }}"
	const expectLastLine = `{{- end }}`
	template = strings.TrimSpace(template)
	lines := strings.Split(template, "\n")
	if len(lines) < 3 {
		err = fmt.Errorf("template file %s needs to contain at least 3 lines", filename)
		return
	}

	firstLine := strings.TrimSpace(lines[0])
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	manifestLines := bytes.Buffer{}
	i := 0
	for _, line := range lines {
		i++
		if i == 1 || i == len(lines) {
			continue
		}
		manifestLines.WriteString(line)
		manifestLines.WriteString("\n")
	}

	if firstLine != expectFirstLine {
		err = fmt.Errorf("first line of template (%s) not in expected format.\nExpected:\n%s\nGot:\n%s", filename, expectFirstLine, firstLine)
		return
	}
	if lastLine != expectLastLine {
		err = fmt.Errorf("last line of template (%s) not in expected format.\nExpected:\n%s\nGot:\n%s", filename, expectLastLine, lastLine)
		return
	}
	err = yaml.Unmarshal(manifestLines.Bytes(), &sealedSecret)
	if err != nil {
		err = fmt.Errorf("template (with first and last line removed) does not contain valid YAML (%s):\n%s", err, manifestLines.String())
		return
	}

	return
}

func main() {
	if len(os.Args) != 3 || os.Args[1] != "rotate" || len(os.Args[2]) == 0 {
		fmt.Printf("Usage:\n\tkubesealplus rotate sealedsecret-filename.yaml\n")
		os.Exit(1)
	}
	filename := os.Args[2]

	secretName, environment, err := nameAndEnvFromFilename(filename)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	_ = secretName

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Cannot open file: %s\n", filename)
		os.Exit(1)
	}

	template, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Cannot read file: %s\n", filename)
		os.Exit(1)
	}

	sealedSecret, err := sealedSecretFromTemplate(filename, environment, string(template))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	keys := []string{}
	for k := range sealedSecret.Spec.EncryptedData {
		keys = append(keys, k)
	}

	// TODO
	secrets := map[string]string{}
	for {
		secrets, err := PromptSecrets(keys, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		redo, err := PromptConfirm(secrets, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		if redo == "" {
			break
		}
	}
	fmt.Printf("Secrets:\n%+v\nError:\n%s", secrets, err)

}

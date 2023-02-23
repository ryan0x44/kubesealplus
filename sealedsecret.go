package main

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

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

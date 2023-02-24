package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func createSealedSecret(secretYAML string, certFilename string) (manifest string, err error) {
	ctx, timeout := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer timeout()
	cmd := exec.CommandContext(ctx, "kubeseal", "-o", "yaml", "--cert", certFilename)
	var stdout, stderr bytes.Buffer
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	io.WriteString(stdin, secretYAML)
	stdin.Close()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return
	}
	if strings.TrimSpace(stderr.String()) != "" {
		err = fmt.Errorf("%s", stderr.String())
	}
	manifest = stdout.String()
	return
}

const firstLineTemplate = "{{- if eq .Values.environment \"%s\" }}"
const lastLineTemplate = `{{- end }}`

func sealedSecretFromTemplate(filename string, environment string, template string) (sealedSecret SealedSecret, err error) {
	expectFirstLine := fmt.Sprintf(firstLineTemplate, environment)
	const expectLastLine = lastLineTemplate
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

func sealedSecretToTemplate(f *os.File, environment string, template string) (out bytes.Buffer, err error) {
	f.Truncate(0)
	f.Seek(0, io.SeekStart)
	for _, writer := range []io.StringWriter{f, &out} {
		writer.WriteString(fmt.Sprintf(firstLineTemplate+"\n", environment))
		writer.WriteString(template)
		writer.WriteString(fmt.Sprintf(lastLineTemplate + "\n"))
	}
	return
}

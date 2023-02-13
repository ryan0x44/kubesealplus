package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

const TPL_FIRST_LINE_REGEX = `^{{- if eq \.Values\.environment "([a-z-]+)" }}$`
const TPL_LAST_LINE = `{{- end }}`

func main() {
	if len(os.Args) != 3 || os.Args[1] != "rotate" || len(os.Args[2]) == 0 {
		fmt.Printf("Usage:\n\tkubesealplus rotate sealedsecret-filename.yaml\n")
		os.Exit(1)
	}
	filename := os.Args[2]

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

	scanner := bufio.NewScanner(bytes.NewReader(template))
	var manifestBuffer bytes.Buffer
	lineCount := 0
	firstMatch := false
	lastMatch := 0
	firstLineRegex, _ := regexp.Compile(TPL_FIRST_LINE_REGEX)
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		if lineCount == 1 {
			if firstMatch = firstLineRegex.MatchString(line); !firstMatch {
				break
			}
		} else if line == TPL_LAST_LINE {
			lastMatch = lineCount
		}
		if lineCount != 1 && lineCount != lastMatch {
			manifestBuffer.WriteString(line)
			manifestBuffer.WriteString("\n")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Cannot read file: %s\n", filename)
		os.Exit(1)
	}
	if !firstMatch {
		fmt.Printf("First line of template (%s) not in expected format.\n", filename)
		os.Exit(1)
	}
	if lineCount != lastMatch {
		fmt.Printf("Last line of template (%s) not in expected format.\n", filename)
		os.Exit(1)
	}

	var sealedSecret SealedSecret
	err = yaml.Unmarshal(manifestBuffer.Bytes(), &sealedSecret)
	if err != nil {
		log.Fatalf("Template without first and last line is not a valid YAML manifest: %s", err)
	}

	keys := []string{}
	for k := range sealedSecret.Spec.EncryptedData {
		keys = append(keys, k)
	}

	// TODO
	fmt.Printf("%s", keys)

}

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const ANSI_ESCAPE_CLEAR = "\033[H\033[2J"

type PromptSecrets struct {
	secrets []PromptSecretInput
}

type PromptSecretInput struct {
	key           string
	kind          PromptSecretInput_Kind
	value         string
	valueFromFile string
}

type PromptSecretInput_Kind string

const PromptSecretInput_Kind_File PromptSecretInput_Kind = "file"
const PromptSecretInput_Kind_String PromptSecretInput_Kind = "string"
const PromptSecretInput_Kind_None PromptSecretInput_Kind = "none"

func (s *PromptSecrets) InitKeys(keys []string) {
	s.secrets = []PromptSecretInput{}
	for _, k := range keys {
		s.secrets = append(s.secrets, PromptSecretInput{
			key: k,
		})
	}
}

func (s PromptSecrets) ToValues() map[string]string {
	values := map[string]string{}
	for _, s := range s.secrets {
		if s.kind == PromptSecretInput_Kind_None {
			continue
		}
		if s.kind == PromptSecretInput_Kind_File {
			values[s.key] = s.valueFromFile
		} else {
			values[s.key] = s.value
		}
	}
	return values
}

func (s *PromptSecrets) Enter(redo int, input io.Reader, output io.Writer) (err error) {
	fmt.Fprintf(
		output,
		"%s%s\n\nPlease enter your secrets for each key then press enter:\n",
		ANSI_ESCAPE_CLEAR,
		strings.Repeat(`-`, 80),
	)
	reader := bufio.NewReader(input)
	for i, secret := range s.secrets {
		if redo > 0 && redo != (i+1) {
			continue
		}
		fmt.Fprintf(output, "%s=", secret.key)
		var value string
		value, err = reader.ReadString('\n')
		if err != nil {
			return
		}
		value = strings.TrimSuffix(value, "\n")
		s.secrets[i] = PromptSecretInput{
			key:   secret.key,
			value: strings.TrimSpace(value),
		}
		valueFromFile, readFileErr := os.ReadFile(value)
		if value == "" {
			s.secrets[i].kind = PromptSecretInput_Kind_None
		} else if readFileErr == nil {
			s.secrets[i].kind = PromptSecretInput_Kind_File
			s.secrets[i].valueFromFile = string(valueFromFile)
		} else {
			s.secrets[i].kind = PromptSecretInput_Kind_String
		}
	}
	return
}

func (s *PromptSecrets) Confirm(input io.Reader, output io.Writer) (redo int, err error) {
	reader := bufio.NewReader(input)
	fmt.Fprintf(
		output,
		"%s%s\n\nPlease review each secret is correct:\n",
		ANSI_ESCAPE_CLEAR,
		strings.Repeat(`-`, 80),
	)
	for i, s := range s.secrets {
		switch s.kind {
		case PromptSecretInput_Kind_File:
			fmt.Fprintf(output, "%d. %s will contain the contents of file %s\n", i+1, s.key, s.value)
		case PromptSecretInput_Kind_None:
			fmt.Fprintf(output, "%d. %s will remain unchanged\n", i+1, s.key)
		case PromptSecretInput_Kind_String:
			fallthrough
		default:
			fmt.Fprintf(output, "%d. %s=%s\n", i+1, s.key, s.value)
		}
	}
	for {
		numString := fmt.Sprintf("1-%d", len(s.secrets))
		if numString == "1-1" {
			numString = "1"
		}
		fmt.Fprintf(output, "\nEnter the secret number to change the value, or Y to confirm\n%s or Y: ", numString)
		in, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		in = strings.TrimSpace(in)
		if in == "Y" {
			break
		}
		redo, err = strconv.Atoi(in)
		if err == nil && redo >= 1 && redo <= len(s.secrets) {
			break
		}
		redo = 0
		fmt.Fprintf(output, "ERROR: Input invalid, please retry.\n")
	}
	fmt.Fprint(output, ANSI_ESCAPE_CLEAR)
	return
}

func PromptClear(output io.Writer) {
	fmt.Fprint(output, ANSI_ESCAPE_CLEAR)
}

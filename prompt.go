package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func PromptSecrets(keys []string, input io.Reader, output io.Writer) (secrets map[string]string, err error) {
	secrets = map[string]string{}
	reader := bufio.NewReader(input)
	for _, key := range keys {
		fmt.Fprintf(output, "%s=", key)
		var value string
		value, err = reader.ReadString('\n')
		if err != nil {
			return
		}
		secrets[key] = strings.TrimSpace(value)
	}
	return
}

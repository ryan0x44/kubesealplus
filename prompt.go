package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
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

func PromptConfirm(secrets map[string]string, input io.Reader, output io.Writer) (redo string, err error) {
	reader := bufio.NewReader(input)
	fmt.Fprintf(output, "\n%s\n\nPlease review each secret is correct:\n", strings.Repeat(`-`, 80))
	i := 0
	var keys []string
	for k, v := range secrets {
		i++
		keys = append(keys, k)
		fmt.Fprintf(output, "%d. %s=%s\n", i, k, v)
	}
	for {
		numString := fmt.Sprintf("1-%d", i)
		if i == 1 {
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
		num, err := strconv.Atoi(in)
		if err == nil && num >= 1 && num <= i {
			redo = keys[num-1]
			break
		}
		fmt.Fprintf(output, "ERROR: Input invalid, please retry.\n")
	}
	return
}

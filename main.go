package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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

	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Cannot open file: %s\n", filename)
		os.Exit(1)
	}
	defer file.Close()

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

	var secrets map[string]string
	var redoSecrets map[string]string
	var redo string
	for {
		if redo != "" {
			redoSecrets, err = PromptSecrets([]string{redo}, os.Stdin, os.Stdout)
		} else {
			secrets, err = PromptSecrets(keys, os.Stdin, os.Stdout)
		}
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		if redo != "" {
			secrets[redo] = redoSecrets[redo]
		}
		redo, err = PromptConfirm(secrets, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		if redo == "" {
			break
		}
	}

	// TODO: allow re-use of existing values

	// TODO: parse filenames and read file contents

	secretYAML, err := createSecretYAML(
		sealedSecret.Metadata.Name,
		sealedSecret.Metadata.Namespace,
		time.Now(),
		secrets,
	)
	if err != nil {
		fmt.Printf("error creating Secret:\n%s\n", err)
		os.Exit(1)
	}

	sealedSecretYAML, err := createSealedSecret(secretYAML)
	if err != nil {
		fmt.Printf("error creating SealedSecret via kubeseal:\n%s\n", err)
		os.Exit(1)
	}

	PromptClear(os.Stdout)

	out, err := sealedSecretToTemplate(file, environment, sealedSecretYAML)
	if err != nil {
		fmt.Printf("error writing SealedSecret to template %s:\n%s\n", filename, err)
		os.Exit(1)
	}
	fmt.Printf("Wrote new SealedSecret to file\n%s\nwith content:\n%s", filename, out.String())

}

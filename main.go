package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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
	command := ""
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}
	switch command {
	case "rotate":
		if len(os.Args) != 3 || len(os.Args[2]) == 0 {
			fmt.Printf("Usage:\n\tkubesealplus rotate (sealedsecret-filename.yaml)\n")
			os.Exit(1)
		}
		rotate(os.Args[2])
	case "config":
		if len(os.Args) != 5 || len(os.Args[2]) == 0 || os.Args[3] != "cert" {
			fmt.Printf("Usage:\n\tkubesealplus config (environment) cert (file path or URL)\n")
			os.Exit(1)
		}
		configure(os.Args[2], os.Args[3], os.Args[4])
	default:
		fmt.Println("Usage: kubesealplus COMMAND")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("\trotate (sealedsecret-filename.yaml)")
		fmt.Println("\tconfig (environment) cert (file path or URL)")
	}
}

var isValidEnv = regexp.MustCompile(`^[a-z0-9-]+$`).MatchString

func configure(environment string, configKey string, configValue string) {
	if !isValidEnv(environment) {
		fmt.Printf("Invalid environment value: %s\n", environment)
		os.Exit(1)
	}

	if configKey != "cert" {
		fmt.Printf("cert is the only config key supported currently")
		os.Exit(1)
	}

	_, err := CertLoad(configValue)
	if err != nil {
		fmt.Printf("Unable to load cert '%s':\n%s\n", configValue, err)
		os.Exit(1)
	}

	configFile, err := ConfigFileDefaultPath("")
	if err != nil {
		fmt.Printf("Unable to determine default config file path: %s", err)
		os.Exit(1)
	}
	configDoc := ConfigDoc{}
	if configDoc.Exists(configFile) {
		err := configDoc.Load(configFile)
		if err != nil {
			fmt.Printf("Unable to load config file: %s", err)
			os.Exit(1)
		}
	}
	configDoc.SetEnvironment(environment, "cert", configValue)
	err = configDoc.Save(configFile)
	if err != nil {
		fmt.Printf("Unable to save config file: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Cert value '%s'\nfor environment '%s'\nsuccessfully saved to config file '%s'\n", configValue, environment, configFile)
}

func rotate(filename string) {
	_, environment, err := nameAndEnvFromFilename(filename)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	configFile, err := ConfigFileDefaultPath("")
	if err != nil {
		panic(err)
	}
	configDoc := ConfigDoc{}
	if !configDoc.Exists(configFile) {
		fmt.Printf("Config for environment '%s' not found. Run this:\n"+
			"kubesealplus config %s cert (your-cert-file)", environment, environment)
		os.Exit(1)
	}
	err = configDoc.Load(configFile)
	if err != nil {
		fmt.Printf("Error loading config file %s: %s\n", configFile, err)
		os.Exit(1)
	}

	certConfigValue := configDoc.Environments[environment]["cert"]
	cert, err := CertLoad(certConfigValue)
	if err != nil {
		fmt.Printf("Unable to load cert '%s':\n%s\n", certConfigValue, err)
		os.Exit(1)
	}
	certFilename, err := ConfigWriteCert(environment, cert)
	if err != nil {
		fmt.Printf("Unable to write latest cert to disk:\n%s\n", err)
		os.Exit(1)
	}

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

	sealedSecretYAML, err := createSealedSecret(secretYAML, certFilename)
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

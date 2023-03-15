package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	case "new":
		if len(os.Args) != 3 || len(os.Args[2]) == 0 {
			fmt.Printf("Usage:\n\tkubesealplus new (sealedsecret-filename.yaml)\n")
			os.Exit(1)
		}
		new(os.Args[2])
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

func loadConfig(environment string) (certFilename string, err error) {
	configFile, err := ConfigFileDefaultPath("")
	if err != nil {
		panic(err)
	}
	configDoc := ConfigDoc{}
	if !configDoc.Exists(configFile) {
		err = fmt.Errorf("Config for environment '%s' not found. Run this:\n"+
			"kubesealplus config %s cert (your-cert-file)", environment, environment)
		return
	}
	err = configDoc.Load(configFile)
	if err != nil {
		err = fmt.Errorf("Error loading config file %s: %s\n", configFile, err)
		return
	}

	certConfigValue := configDoc.Environments[environment]["cert"]
	// TODO: implement caching of cert load.
	// we probably only need to download it at most once per hour (or day?)
	cert, err := CertLoad(certConfigValue)
	if err != nil {
		err = fmt.Errorf("unable to load cert '%s':\n%s\n", certConfigValue, err)
		return
	}
	certFilename, err = ConfigWriteCert(environment, cert)
	if err != nil {
		err = fmt.Errorf("Unable to write latest cert to disk:\n%s\n", err)
		return
	}
	return
}

func new(filename string) {
	fileInfo, err := os.Stat(filename)
	if err == nil && fileInfo != nil {
		fmt.Printf("Error: cannot create new file as file already exists\n\t%s\n", filename)
		os.Exit(1)
	}

	secretName, environment, err := nameAndEnvFromFilename(filename)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	secrets := PromptSecrets{}
	namespace, err := secrets.Namespace(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	sealedSecret := SealedSecret{Environment: environment}
	sealedSecret.Init(secretName, namespace)

	rotateAndNew(&sealedSecret, secrets)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating new file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()
	out, err := sealedSecret.ToTemplate(file, environment)
	if err != nil {
		fmt.Printf("error writing SealedSecret file %s: %s\n", filename, err)
		os.Exit(1)
	}
	fmt.Printf("Created SealedSecret file '%s' with content:\n%s", filename, out.String())
}

func rotate(filename string) {
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
	_, environment, err := nameAndEnvFromFilename(filename)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	sealedSecret, err := sealedSecretFromTemplate(filename, environment, string(template))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	secrets := PromptSecrets{}
	for k := range sealedSecret.Spec.EncryptedData {
		secrets.InitKey(k)
	}
	rotateAndNew(&sealedSecret, secrets)

	out, err := sealedSecret.ToTemplate(file, environment)
	if err != nil {
		fmt.Printf("error writing SealedSecret file %s: %s\n", filename, err)
		os.Exit(1)
	}
	fmt.Printf("Updated SealedSecret file '%s' with content:\n%s", filename, out.String())
}

func rotateAndNew(sealedSecret *SealedSecret, secrets PromptSecrets) {
	certFilename, err := loadConfig(sealedSecret.Environment)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	redo := 0
	for {
		var err error
		if len(secrets.secrets) > 0 {
			err = secrets.Update(redo, os.Stdin, os.Stdout)
		} else {
			err = secrets.Enter(os.Stdin, os.Stdout)
		}
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		redo, err = secrets.Confirm(os.Stdin, os.Stdout)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		if redo == 0 {
			break
		}
	}
	PromptClear(os.Stdout)

	newSecrets := secrets.ToValues()
	secretYAML, err := createSecretYAML(
		sealedSecret.Spec.Template.Metadata,
		newSecrets,
	)
	if err != nil {
		fmt.Printf("error creating Secret:\n%s\n", err)
		os.Exit(1)
	}

	newSealedSecrets, err := createSealedSecrets(secretYAML, certFilename)
	if err != nil {
		fmt.Printf("error creating SealedSecret via kubeseal:\n%s\n", err)
		os.Exit(1)
	}
	if len(newSealedSecrets) != len(newSecrets) {
		fmt.Printf("error creating SealedSecret via kubeseal:\n%s\n",
			"number of secrets returned do not match number given")
		os.Exit(1)
	}
	if sealedSecret.Spec.EncryptedData == nil {
		sealedSecret.Spec.EncryptedData = map[string]string{}
	}
	for k, v := range newSealedSecrets {
		sealedSecret.Spec.EncryptedData[k] = v
	}
}

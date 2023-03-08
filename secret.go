package main

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

type secretManifest struct {
	ApiVersion string             `yaml:"apiVersion"`
	Kind       string             `yaml:"kind"`
	Type       string             `yaml:"type"`
	Data       map[string]string  `yaml:"data"`
	Metadata   map[string]*string `yaml:"metadata"`
}

func createSecretYAML(
	metadata map[string]*string,
	secrets map[string]string,
) (manifestYAML string, err error) {
	manifest := secretManifest{
		ApiVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		Data:       map[string]string{},
		Metadata:   metadata,
	}
	for k, v := range secrets {
		manifest.Data[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}
	manifestYAMLBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return
	}
	manifestYAML = string(manifestYAMLBytes)
	return
}

func validateSecretName(name string) error {
	// https://kubernetes.io/docs/concepts/configuration/secret/#restriction-names-data
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names
	if len(name) > 63 {
		return fmt.Errorf("secret name cannot exceed 63 characters per RFC 1123")
	}
	match, err := regexp.MatchString("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", name)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf(
			"%s.\n- %s\n- %s\n- %s",
			"secret name must be valid DNS subdomain",
			"must only contain lowercase alphanumeric characters or -",
			"must start with a lowercase alphanumeric character",
			"must end with a lowercase alphanumeric character",
		)
	}

	return nil
}

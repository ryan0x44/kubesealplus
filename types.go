package main

import "time"

type SealedSecret struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty"`
		Name              string    `yaml:"name"`
		Namespace         string    `yaml:"namespace,omitempty"`
	} `yaml:"metadata"`
	Spec struct {
		EncryptedData map[string]string `yaml:"encryptedData,omitempty"`
		Template      struct {
			Data     *map[string]string `yaml:"data"`
			Metadata map[string]string  `yaml:"metadata"`
		} `yaml:"template"`
	} `yaml:"spec`
}

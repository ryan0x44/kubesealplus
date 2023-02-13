package main

import "time"

type SealedSecret struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		CreationTimestamp time.Time `json:"creationTimestamp,omitempty"`
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace,omitempty"`
	} `json:"metadata"`
	Spec struct {
		EncryptedData map[string]string `json:"encryptedData,omitempty"`
		Template      struct {
			Data     map[string]string `json:"data"`
			Metadata map[string]string `json:"metadata"`
		} `json:"template"`
	} `json:"spec`
}

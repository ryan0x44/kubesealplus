package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConfigDoc struct {
	Environments map[string]map[string]string `yaml:"environments"`
}

func (doc *ConfigDoc) SetEnvironment(environment string, key string, value string) {
	if doc.Environments == nil {
		doc.Environments = map[string]map[string]string{}
	}
	if _, exists := doc.Environments[environment]; !exists {
		doc.Environments[environment] = map[string]string{}
	}
	doc.Environments[environment][key] = value
}

func (doc *ConfigDoc) Exists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return true
}

func (doc *ConfigDoc) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open config file '%s': %s", filename, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read file '%s': %s", filename, err)
	}

	err = yaml.Unmarshal(content, &doc)
	if err != nil {
		return fmt.Errorf("cannot parse YAML in file '%s': %s", filename, err)
	}

	return nil
}

func (doc *ConfigDoc) Save(filename string) error {
	filedir := filepath.Dir(filename)
	_, err := os.Stat(filedir)
	if err != nil {
		err = os.Mkdir(filedir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create/open config file '%s': %s", filename, err)
	}
	defer file.Close()

	content, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("cannot marshal YAML: %s", err)
	}

	bytesWritten, err := io.WriteString(file, string(content))
	if err == nil && bytesWritten != len(content) {
		return fmt.Errorf("failed to write all bytes to config file '%s'", filename)
	}
	if err != nil {
		return fmt.Errorf("cannot write to config file '%s': %s", filename, err)
	}

	return nil
}

func ConfigDirDefaultPath() (path string, err error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path = fmt.Sprintf("%s/.kubesealplus", dirname)
	return
}

func ConfigFileDefaultPath(filename string) (path string, err error) {
	if filename == "" {
		filename = "config.yaml"
	}
	dirname, err := ConfigDirDefaultPath()
	if err != nil {
		return
	}
	path = fmt.Sprintf("%s/%s", dirname, filename)
	return
}

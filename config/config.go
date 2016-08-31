package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	autogradRootEnvKey = "AUTOGRAD_ROOT"
	configFileName     = "configuration.yml"
)

func GetAutogradRoot() (string, error) {
	root := os.Getenv(autogradRootEnvKey)
	if root == "" {
		return "", fmt.Errorf("%s environment variable not set", autogradRootEnvKey)
	}
	return root, nil
}

func Load(autogradRoot string) (*Config, error) {
	file, err := ioutil.ReadFile(filepath.Join(autogradRoot, configFileName))
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

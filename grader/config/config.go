package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	configFileName = "configuration.yml"
)

type Config struct {
	Grader GraderConfig `yaml:"grader"`
}

type GraderConfig struct {
	InitCommands    [][]string `yaml:"init_commands"`
	SetupCommands   [][]string `yaml:"setup_commands"`
	GradeCommand    []string   `yaml:"grade_command"`
	CleanupCommands [][]string `yaml:"cleanup_commands"`
	GradeTimeout    int        `yaml:"grade_timeout"`
}

func Load(graderRoot string) (*Config, error) {
	file, err := ioutil.ReadFile(filepath.Join(graderRoot, configFileName))
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

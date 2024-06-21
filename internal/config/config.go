package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Jobs []struct {
		Name   string `yaml:"name"`
		Cron   string `yaml:"cron"`
		Script string `yaml:"script"`
	} `yaml:"jobs"`
}

func LoadConfig(configPath string) (Config, error) {
	var config Config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

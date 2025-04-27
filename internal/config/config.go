package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int      `yaml:"port"`
	Backends []string `yaml:"backends"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

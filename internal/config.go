package internal

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Resources []Resource `yaml:"resources"`
}

type Resource struct {
	Service   string  `yaml:"service"`
	Method    string  `yaml:"method"`
	Path      string  `yaml:"path"`
	Algorithm string  `yaml:"algorithm"`
	Rules     RuleSet `yaml:"rules"`
}

type RuleSet struct {
	Limits []Limit `yaml:"limits"`
}

type Limit struct {
	WindowSeconds int `yaml:"window_seconds"`
	Limit         int `yaml:"limit"`
}

func LoadConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

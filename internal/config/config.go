package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level driftwatch configuration.
type Config struct {
	Version  string    `yaml:"version"`
	Services []Service `yaml:"services"`
}

// Service represents a single service entry to watch for drift.
type Service struct {
	Name       string            `yaml:"name"`
	Repository string            `yaml:"repository"`
	Branch     string            `yaml:"branch"`
	Manifest   string            `yaml:"manifest"`
	Labels     map[string]string `yaml:"labels,omitempty"`
}

// Load reads and parses a driftwatch config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate performs basic sanity checks on the loaded configuration.
func (c *Config) validate() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("no services defined")
	}

	seen := make(map[string]struct{}, len(c.Services))
	for i, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("service[%d]: name is required", i)
		}
		if svc.Repository == "" {
			return fmt.Errorf("service %q: repository is required", svc.Name)
		}
		if svc.Manifest == "" {
			return fmt.Errorf("service %q: manifest path is required", svc.Name)
		}
		if _, dup := seen[svc.Name]; dup {
			return fmt.Errorf("duplicate service name %q", svc.Name)
		}
		seen[svc.Name] = struct{}{}
	}
	return nil
}

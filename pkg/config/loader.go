package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads an agent configuration from a YAML file
func Load(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg AgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadOrDefault loads config from path, or returns default for the agent if path is empty
func LoadOrDefault(path, agentName string) (*AgentConfig, error) {
	if path != "" {
		return Load(path)
	}
	return DefaultConfig(agentName), nil
}

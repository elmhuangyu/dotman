package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// ModuleConfig represents the structure of a Dotfile configuration
type ModuleConfig struct {
	Dir       string
	TargetDir string   `yaml:"target_dir"`
	Ignores   []string `yaml:"ignores"`
}

// LoadConfig loads and parses a Dotfile configuration from the specified directory
func LoadConfig(moduleDir string) (*ModuleConfig, error) {
	configPath := filepath.Join(moduleDir, "Dotfile")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // No config file is not an error
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML
	var config ModuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Validate config
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid config in %s: %w", configPath, err)
	}

	config.Dir = moduleDir

	return &config, nil
}

// validate validates the configuration structure and values
func (config *ModuleConfig) validate() error {
	if config.TargetDir == "" {
		return fmt.Errorf("target_dir field is required")
	}

	// target_dir must be an absolute path
	if !filepath.IsAbs(config.TargetDir) {
		return fmt.Errorf("target_dir must be an absolute path")
	}

	// For absolute paths, ensure they're properly formatted
	if filepath.Clean(config.TargetDir) != config.TargetDir {
		return fmt.Errorf("target_dir contains invalid path components")
	}

	// Validate ignores list - ensure no empty strings
	for i, ignore := range config.Ignores {
		if ignore == "" {
			return fmt.Errorf("ignores[%d] cannot be empty", i)
		}
	}

	return nil
}

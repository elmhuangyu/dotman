package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/goccy/go-yaml"
)

// RootConfig represents the root configuration structure
type RootConfig struct {
	Vars           map[string]string `yaml:"vars"`
	ExcludeModules []string          `yaml:"exclude_modules"`
}

// LoadRootConfig loads and parses a root configuration from the specified directory
func LoadRootConfig(dir string) (RootConfig, error) {
	configPath := filepath.Join(dir, "DotRoot")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return RootConfig{}, nil // No config file is not an error
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return RootConfig{}, fmt.Errorf("failed to read root config file %s: %w", configPath, err)
	}

	// Parse YAML
	var config RootConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return RootConfig{}, fmt.Errorf("failed to parse root config file %s: %w", configPath, err)
	}

	// Validate config
	if err := config.validate(); err != nil {
		return RootConfig{}, fmt.Errorf("invalid root config in %s: %w", configPath, err)
	}

	return config, nil
}

// validate validates the root configuration structure and values
func (config *RootConfig) validate() error {
	// Validate vars keys - only alphanumeric characters allowed
	varKeyPattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	for key := range config.Vars {
		if !varKeyPattern.MatchString(key) {
			return fmt.Errorf("vars key '%s' contains invalid characters, only a-zA-Z0-9 are allowed", key)
		}
	}

	// Validate exclude_modules strings - alphanumeric, hyphen, underscore, and dot allowed
	excludeModulePattern := regexp.MustCompile(`^[-_\.a-zA-Z0-9]+$`)
	for i, module := range config.ExcludeModules {
		if module == "" {
			return fmt.Errorf("exclude_modules[%d] cannot be empty", i)
		}
		if !excludeModulePattern.MatchString(module) {
			return fmt.Errorf("exclude_modules[%d] '%s' contains invalid characters, only -_\\.a-zA-Z0-9 are allowed", i, module)
		}
	}

	return nil
}

// IsModuleExcluded checks if a module name is in the exclude list
func (config *RootConfig) IsModuleExcluded(moduleName string) bool {
	for _, excludeModule := range config.ExcludeModules {
		if moduleName == excludeModule {
			return true
		}
	}
	return false
}

package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	RootConfig RootConfig
	Modules    []ModuleConfig
}

func LoadDir(rootDir string) (*Config, error) {
	// Load root config
	rootConfig, err := LoadRootConfig(rootDir)
	if err != nil {
		return nil, err
	}

	ls, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	var modules []ModuleConfig
	for _, entry := range ls {
		if !entry.IsDir() {
			continue
		}

		// Skip excluded modules
		if rootConfig.IsModuleExcluded(entry.Name()) {
			continue
		}

		moduleDir := filepath.Join(rootDir, entry.Name())
		moduleConfig, err := LoadConfig(moduleDir)
		if err != nil {
			return nil, err
		}
		if moduleConfig != nil {
			modules = append(modules, *moduleConfig)
		}
	}

	return &Config{
		RootConfig: rootConfig,
		Modules:    modules,
	}, nil
}

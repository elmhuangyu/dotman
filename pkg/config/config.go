package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Modules []ModuleConfig
}

func LoadDir(rootDir string) (*Config, error) {
	ls, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	var modules []ModuleConfig
	for _, entry := range ls {
		if !entry.IsDir() {
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

	return &Config{Modules: modules}, nil
}

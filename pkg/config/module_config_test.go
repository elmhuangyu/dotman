package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		configContent string
		setupFunc     func(t *testing.T, dir string) string
		wantConfig    *ModuleConfig
		wantErr       bool
		errContains   string
	}{
		{
			name:          "ValidConfig",
			configContent: `target_dir: "/home/user/.config/nvim"`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "Dotfile")
				err := os.WriteFile(configPath, []byte(`target_dir: "/home/user/.config/nvim"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig: &ModuleConfig{
				Dir:       filepath.Join(tmpDir, "ValidConfig"),
				TargetDir: "/home/user/.config/nvim",
				Ignores:   nil,
			},
			wantErr: false,
		},
		{
			name:          "NoConfigFile",
			configContent: "",
			setupFunc: func(t *testing.T, dir string) string {
				emptyDir := filepath.Join(dir, "empty")
				err := os.Mkdir(emptyDir, 0755)
				require.NoError(t, err)
				return emptyDir
			},
			wantConfig: nil,
			wantErr:    false,
		},
		{
			name:          "InvalidYAML",
			configContent: `target_dir: "/home/user/.config/nvim`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "Dotfile")
				err := os.WriteFile(configPath, []byte(`target_dir: "/home/user/.config/nvim`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig:  nil,
			wantErr:     true,
			errContains: "failed to parse config file",
		},
		{
			name:          "MissingTargetDir",
			configContent: `other_field: "value"`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "Dotfile")
				err := os.WriteFile(configPath, []byte(`other_field: "value"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig:  nil,
			wantErr:     true,
			errContains: "target_dir field is required",
		},
		{
			name:          "RelativePathTargetDir",
			configContent: `target_dir: ".config/nvim"`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "Dotfile")
				err := os.WriteFile(configPath, []byte(`target_dir: ".config/nvim"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig:  nil,
			wantErr:     true,
			errContains: "target_dir must be an absolute path",
		},
		{
			name: "ValidConfigWithIgnores",
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "Dotfile")
				err := os.WriteFile(configPath, []byte(`target_dir: "/home/user/.config/nvim"
ignores:
  - "*.log"
  - "temp/"
  - ".env"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig: &ModuleConfig{
				Dir:       filepath.Join(tmpDir, "ValidConfigWithIgnores"),
				TargetDir: "/home/user/.config/nvim",
				Ignores:   []string{"*.log", "temp/", ".env"},
			},
			wantErr: false,
		},
		{
			name: "ValidConfigWithEmptyIgnores",
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "Dotfile")
				err := os.WriteFile(configPath, []byte(`target_dir: "/home/user/.config/nvim"
ignores: []`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig: &ModuleConfig{
				Dir:       filepath.Join(tmpDir, "ValidConfigWithEmptyIgnores"),
				TargetDir: "/home/user/.config/nvim",
				Ignores:   []string{},
			},
			wantErr: false,
		},
		{
			name: "InvalidIgnoresWithEmptyString",
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "Dotfile")
				err := os.WriteFile(configPath, []byte(`target_dir: "/home/user/.config/nvim"
ignores:
  - "*.log"
  - ""
  - "temp/"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig:  nil,
			wantErr:     true,
			errContains: "ignores[1] cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.Mkdir(testDir, 0755)
			require.NoError(t, err)

			configDir := tt.setupFunc(t, testDir)

			config, err := LoadConfig(configDir)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantConfig, config)
			}
		})
	}
}

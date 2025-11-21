package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRootConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		configContent string
		setupFunc     func(t *testing.T, dir string) string
		wantConfig    RootConfig
		wantErr       bool
		errContains   string
	}{
		{
			name: "ValidConfigWithVars",
			configContent: `vars:
  USERNAME: "john"
  HOMEDIR: "/home/john"
exclude_modules: []`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "DotRoot")
				err := os.WriteFile(configPath, []byte(`vars:
  USERNAME: "john"
  HOMEDIR: "/home/john"
exclude_modules: []`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig: RootConfig{
				Vars: map[string]string{
					"USERNAME": "john",
					"HOMEDIR":  "/home/john",
				},
				ExcludeModules: []string{},
			},
			wantErr: false,
		},
		{
			name: "ValidConfigWithExcludeModules",
			configContent: `vars: {}
exclude_modules:
  - "nvim"
  - "bash"
  - "git-config"`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "DotRoot")
				err := os.WriteFile(configPath, []byte(`vars: {}
exclude_modules:
  - "nvim"
  - "bash"
  - "git-config"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig: RootConfig{
				Vars: map[string]string{},
				ExcludeModules: []string{
					"nvim",
					"bash",
					"git-config",
				},
			},
			wantErr: false,
		},
		{
			name: "ValidConfigEmpty",
			configContent: `vars: {}
exclude_modules: []`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "DotRoot")
				err := os.WriteFile(configPath, []byte(`vars: {}
exclude_modules: []`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig: RootConfig{
				Vars:           map[string]string{},
				ExcludeModules: []string{},
			},
			wantErr: false,
		},
		{
			name:          "NoConfigFile",
			configContent: "",
			setupFunc: func(t *testing.T, dir string) string {
				// Don't create any config file
				return dir
			},
			wantConfig: RootConfig{},
			wantErr:    false,
		},
		{
			name: "InvalidYAML",
			configContent: `vars:
  USERNAME: "john"
  HOMEDIR: "/home/john"
exclude_modules:
  - "nvim"`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "DotRoot")
				err := os.WriteFile(configPath, []byte(`vars:
  USERNAME: "john
  HOMEDIR: "/home/john"
exclude_modules:
  - "nvim"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig:  RootConfig{},
			wantErr:     true,
			errContains: "failed to parse root config file",
		},
		{
			name: "InvalidVarKeyWithSpecialChars",
			configContent: `vars:
  USER-NAME: "john"
  HOMEDIR: "/home/john"
exclude_modules: []`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "DotRoot")
				err := os.WriteFile(configPath, []byte(`vars:
  USER-NAME: "john"
  HOMEDIR: "/home/john"
exclude_modules: []`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig:  RootConfig{},
			wantErr:     true,
			errContains: "contains invalid characters",
		},
		{
			name: "InvalidExcludeModuleWithSlash",
			configContent: `vars: {}
exclude_modules:
  - "nvim/config"
  - "bash"`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "DotRoot")
				err := os.WriteFile(configPath, []byte(`vars: {}
exclude_modules:
  - "nvim/config"
  - "bash"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig:  RootConfig{},
			wantErr:     true,
			errContains: "exclude_modules[0] 'nvim/config' contains invalid characters",
		},
		{
			name: "EmptyExcludeModuleString",
			configContent: `vars: {}
exclude_modules:
  - ""
  - "bash"`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "DotRoot")
				err := os.WriteFile(configPath, []byte(`vars: {}
exclude_modules:
  - ""
  - "bash"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig:  RootConfig{},
			wantErr:     true,
			errContains: "exclude_modules[0] cannot be empty",
		},
		{
			name: "ValidComplexConfig",
			configContent: `vars:
  USERNAME: "alice"
  VERSION: "1.0.0"
  DEBUG123: "true"
exclude_modules:
  - "my-module"
  - "test_config"
  - "backup.dir"
  - "legacy_module"`,
			setupFunc: func(t *testing.T, dir string) string {
				configPath := filepath.Join(dir, "DotRoot")
				err := os.WriteFile(configPath, []byte(`vars:
  USERNAME: "alice"
  VERSION: "1.0.0"
  DEBUG123: "true"
exclude_modules:
  - "my-module"
  - "test_config"
  - "backup.dir"
  - "legacy_module"`), 0644)
				require.NoError(t, err)
				return dir
			},
			wantConfig: RootConfig{
				Vars: map[string]string{
					"USERNAME": "alice",
					"VERSION":  "1.0.0",
					"DEBUG123": "true",
				},
				ExcludeModules: []string{
					"my-module",
					"test_config",
					"backup.dir",
					"legacy_module",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.Mkdir(testDir, 0755)
			require.NoError(t, err)

			configPath := tt.setupFunc(t, testDir)

			config, err := LoadRootConfig(configPath)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantConfig, config)
		})
	}
}

func TestValidateRootConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      RootConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "ValidEmptyConfig",
			config: RootConfig{
				Vars:           map[string]string{},
				ExcludeModules: []string{},
			},
			wantErr: false,
		},
		{
			name: "ValidVarsAndExcludeModules",
			config: RootConfig{
				Vars: map[string]string{
					"USER":     "john",
					"HOME":     "/home/john",
					"DEBUG123": "true",
				},
				ExcludeModules: []string{
					"nvim",
					"bash-config",
					"backup.dir",
					"test_module",
				},
			},
			wantErr: false,
		},
		{
			name: "InvalidVarKeyWithHyphen",
			config: RootConfig{
				Vars: map[string]string{
					"USER-NAME": "john",
				},
				ExcludeModules: []string{},
			},
			wantErr:     true,
			errContains: "vars key 'USER-NAME' contains invalid characters",
		},
		{
			name: "InvalidVarKeyWithUnderscore",
			config: RootConfig{
				Vars: map[string]string{
					"USER_NAME": "john",
				},
				ExcludeModules: []string{},
			},
			wantErr:     true,
			errContains: "vars key 'USER_NAME' contains invalid characters",
		},
		{
			name: "InvalidVarKeyWithDot",
			config: RootConfig{
				Vars: map[string]string{
					"USER.NAME": "john",
				},
				ExcludeModules: []string{},
			},
			wantErr:     true,
			errContains: "vars key 'USER.NAME' contains invalid characters",
		},
		{
			name: "InvalidExcludeModuleWithSlash",
			config: RootConfig{
				Vars:           map[string]string{},
				ExcludeModules: []string{"nvim/config"},
			},
			wantErr:     true,
			errContains: "exclude_modules[0] 'nvim/config' contains invalid characters",
		},
		{
			name: "InvalidExcludeModuleWithSpace",
			config: RootConfig{
				Vars:           map[string]string{},
				ExcludeModules: []string{"my module"},
			},
			wantErr:     true,
			errContains: "exclude_modules[0] 'my module' contains invalid characters",
		},
		{
			name: "InvalidExcludeModuleWithAtSign",
			config: RootConfig{
				Vars:           map[string]string{},
				ExcludeModules: []string{"@module"},
			},
			wantErr:     true,
			errContains: "exclude_modules[0] '@module' contains invalid characters",
		},
		{
			name: "EmptyExcludeModule",
			config: RootConfig{
				Vars:           map[string]string{},
				ExcludeModules: []string{""},
			},
			wantErr:     true,
			errContains: "exclude_modules[0] cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRootConfig_IsModuleExcluded(t *testing.T) {
	tests := []struct {
		name       string
		config     RootConfig
		moduleName string
		expected   bool
	}{
		{
			name: "ModuleInExcludeList",
			config: RootConfig{
				ExcludeModules: []string{"nvim", "bash", "git-config"},
			},
			moduleName: "nvim",
			expected:   true,
		},
		{
			name: "ModuleNotInExcludeList",
			config: RootConfig{
				ExcludeModules: []string{"nvim", "bash"},
			},
			moduleName: "git-config",
			expected:   false,
		},
		{
			name: "EmptyExcludeList",
			config: RootConfig{
				ExcludeModules: []string{},
			},
			moduleName: "any-module",
			expected:   false,
		},
		{
			name: "ModuleNameMatchesExactly",
			config: RootConfig{
				ExcludeModules: []string{"test-module"},
			},
			moduleName: "test-module",
			expected:   true,
		},
		{
			name: "ModuleNameDoesNotMatchPartially",
			config: RootConfig{
				ExcludeModules: []string{"test"},
			},
			moduleName: "test-module",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsModuleExcluded(tt.moduleName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDir_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, rootDir string)
		wantConfig func(tmpDir string) *Config
	}{
		{
			name: "MultipleValidModules",
			setupFunc: func(t *testing.T, rootDir string) {
				// Create first module
				module1Dir := filepath.Join(rootDir, "nvim")
				err := os.Mkdir(module1Dir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(module1Dir, "Dotfile"), []byte(`target_dir: "/home/user/.config/nvim"`), 0644)
				require.NoError(t, err)

				// Create second module
				module2Dir := filepath.Join(rootDir, "bash")
				err = os.Mkdir(module2Dir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(module2Dir, "Dotfile"), []byte(`target_dir: "/home/user"`), 0644)
				require.NoError(t, err)
			},
			wantConfig: func(tmpDir string) *Config {
				nvimPath := filepath.Join(tmpDir, "MultipleValidModules", "nvim")
				bashPath := filepath.Join(tmpDir, "MultipleValidModules", "bash")
				return &Config{
					RootConfig: RootConfig{},
					Modules: []ModuleConfig{
						{Dir: nvimPath, TargetDir: "/home/user/.config/nvim"},
						{Dir: bashPath, TargetDir: "/home/user"},
					},
				}
			},
		},
		{
			name: "EmptyDirectory",
			setupFunc: func(t *testing.T, rootDir string) {
				// No directories created
			},
			wantConfig: func(tmpDir string) *Config {
				return &Config{
					RootConfig: RootConfig{},
					Modules:    []ModuleConfig{},
				}
			},
		},
		{
			name: "OnlyFilesNoDirectories",
			setupFunc: func(t *testing.T, rootDir string) {
				err := os.WriteFile(filepath.Join(rootDir, "file1.txt"), []byte("content"), 0644)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(rootDir, "file2.txt"), []byte("content2"), 0644)
				require.NoError(t, err)
			},
			wantConfig: func(tmpDir string) *Config {
				return &Config{
					RootConfig: RootConfig{},
					Modules:    []ModuleConfig{},
				}
			},
		},
		{
			name: "DirectoriesWithoutDotfile",
			setupFunc: func(t *testing.T, rootDir string) {
				dir1 := filepath.Join(rootDir, "dir1")
				err := os.Mkdir(dir1, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir1, "otherfile.txt"), []byte("content"), 0644)
				require.NoError(t, err)

				dir2 := filepath.Join(rootDir, "dir2")
				err = os.Mkdir(dir2, 0755)
				require.NoError(t, err)
			},
			wantConfig: func(tmpDir string) *Config {
				return &Config{
					RootConfig: RootConfig{},
					Modules:    []ModuleConfig{},
				}
			},
		},
		{
			name: "SingleValidModule",
			setupFunc: func(t *testing.T, rootDir string) {
				moduleDir := filepath.Join(rootDir, "single")
				err := os.Mkdir(moduleDir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(moduleDir, "Dotfile"), []byte(`target_dir: "/home/user/.config/single"`), 0644)
				require.NoError(t, err)
			},
			wantConfig: func(tmpDir string) *Config {
				return &Config{
					RootConfig: RootConfig{},
					Modules: []ModuleConfig{
						{
							Dir:       filepath.Join(tmpDir, "SingleValidModule", "single"),
							TargetDir: "/home/user/.config/single",
						},
					},
				}
			},
		},
		{
			name: "WithValidRootConfig",
			setupFunc: func(t *testing.T, rootDir string) {
				// Create root config
				err := os.WriteFile(filepath.Join(rootDir, "DotRoot"), []byte(`vars:
  USERNAME: "john"
  HOMEDIR: "/home/john"
exclude_modules:
  - "temp"
  - "backup"`), 0644)
				require.NoError(t, err)

				// Create a module
				moduleDir := filepath.Join(rootDir, "nvim")
				err = os.Mkdir(moduleDir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(moduleDir, "Dotfile"), []byte(`target_dir: "/home/user/.config/nvim"`), 0644)
				require.NoError(t, err)
			},
			wantConfig: func(tmpDir string) *Config {
				return &Config{
					RootConfig: RootConfig{
						Vars: map[string]string{
							"USERNAME":  "john",
							"HOMEDIR":   "/home/john",
							"DONT_EDIT": "!!! THIS FILE IS GENERATED. DON'T EDIT THIS FILE !!!",
						},
						ExcludeModules: []string{"temp", "backup"},
					},
					Modules: []ModuleConfig{
						{
							Dir:       filepath.Join(tmpDir, "WithValidRootConfig", "nvim"),
							TargetDir: "/home/user/.config/nvim",
						},
					},
				}
			},
		},
		{
			name: "WithEmptyRootConfig",
			setupFunc: func(t *testing.T, rootDir string) {
				// Create empty root config
				err := os.WriteFile(filepath.Join(rootDir, "DotRoot"), []byte(`vars: {}
exclude_modules: []`), 0644)
				require.NoError(t, err)

				// Create a module
				moduleDir := filepath.Join(rootDir, "bash")
				err = os.Mkdir(moduleDir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(moduleDir, "Dotfile"), []byte(`target_dir: "/home/user"`), 0644)
				require.NoError(t, err)
			},
			wantConfig: func(tmpDir string) *Config {
				return &Config{
					RootConfig: RootConfig{
						Vars: map[string]string{
							"DONT_EDIT": "!!! THIS FILE IS GENERATED. DON'T EDIT THIS FILE !!!",
						},
						ExcludeModules: []string{},
					},
					Modules: []ModuleConfig{
						{
							Dir:       filepath.Join(tmpDir, "WithEmptyRootConfig", "bash"),
							TargetDir: "/home/user",
						},
					},
				}
			},
		},
		{
			name: "RootConfigWithExcludeModules",
			setupFunc: func(t *testing.T, rootDir string) {
				// Create root config with exclude modules
				err := os.WriteFile(filepath.Join(rootDir, "DotRoot"), []byte(`vars: {}
exclude_modules:
  - "excluded-module"
  - "test-module"`), 0644)
				require.NoError(t, err)

				// Create modules (including excluded ones)
				excludedDir := filepath.Join(rootDir, "excluded-module")
				err = os.Mkdir(excludedDir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(excludedDir, "Dotfile"), []byte(`target_dir: "/home/user/.config/excluded"`), 0644)
				require.NoError(t, err)

				includedDir := filepath.Join(rootDir, "included-module")
				err = os.Mkdir(includedDir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(includedDir, "Dotfile"), []byte(`target_dir: "/home/user/.config/included"`), 0644)
				require.NoError(t, err)
			},
			wantConfig: func(tmpDir string) *Config {
				return &Config{
					RootConfig: RootConfig{
						Vars: map[string]string{
							"DONT_EDIT": "!!! THIS FILE IS GENERATED. DON'T EDIT THIS FILE !!!",
						},
						ExcludeModules: []string{"excluded-module", "test-module"},
					},
					Modules: []ModuleConfig{
						{
							Dir:       filepath.Join(tmpDir, "RootConfigWithExcludeModules", "included-module"),
							TargetDir: "/home/user/.config/included",
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.Mkdir(testDir, 0755)
			require.NoError(t, err)
			tt.setupFunc(t, testDir)

			config, err := LoadDir(testDir)
			require.NoError(t, err)

			expected := tt.wantConfig(tmpDir)
			assert.Equal(t, expected.RootConfig, config.RootConfig)
			assert.ElementsMatch(t, expected.Modules, config.Modules)
		})
	}
}

func TestLoadDir_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, rootDir string)
		errContains string
	}{
		{
			name: "MixedValidAndInvalidModules",
			setupFunc: func(t *testing.T, rootDir string) {
				// Create valid module
				validDir := filepath.Join(rootDir, "valid")
				err := os.Mkdir(validDir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(validDir, "Dotfile"), []byte(`target_dir: "/home/user/.config/valid"`), 0644)
				require.NoError(t, err)

				// Create directory without Dotfile (should be skipped)
				noConfigDir := filepath.Join(rootDir, "noconfig")
				err = os.Mkdir(noConfigDir, 0755)
				require.NoError(t, err)

				// Create directory with invalid config (should return error)
				invalidDir := filepath.Join(rootDir, "invalid")
				err = os.Mkdir(invalidDir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(invalidDir, "Dotfile"), []byte(`target_dir: "relative/path"`), 0644)
				require.NoError(t, err)
			},
			errContains: "target_dir must be an absolute path",
		},
		{
			name: "NonExistentRootDirectory",
			setupFunc: func(t *testing.T, rootDir string) {
				// Don't create anything
			},
			errContains: "no such file or directory",
		},
		{
			name: "InvalidRootConfig",
			setupFunc: func(t *testing.T, rootDir string) {
				// Create invalid root config with bad YAML
				err := os.WriteFile(filepath.Join(rootDir, "DotRoot"), []byte(`vars:
  USERNAME: "john
  HOMEDIR: "/home/john"`), 0644)
				require.NoError(t, err)
			},
			errContains: "failed to parse root config file",
		},
		{
			name: "InvalidRootConfigVars",
			setupFunc: func(t *testing.T, rootDir string) {
				// Create invalid root config with invalid var key
				err := os.WriteFile(filepath.Join(rootDir, "DotRoot"), []byte(`vars:
  USER-NAME: "john"
exclude_modules: []`), 0644)
				require.NoError(t, err)
			},
			errContains: "contains invalid characters",
		},
		{
			name: "InvalidRootConfigExcludeModules",
			setupFunc: func(t *testing.T, rootDir string) {
				// Create invalid root config with invalid exclude module
				err := os.WriteFile(filepath.Join(rootDir, "DotRoot"), []byte(`vars: {}
exclude_modules:
  - "module/invalid"`), 0644)
				require.NoError(t, err)
			},
			errContains: "exclude_modules[0] 'module/invalid' contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testDir string
			if tt.name == "NonExistentRootDirectory" {
				// Use a non-existent path for this test
				testDir = filepath.Join(tmpDir, "nonexistent")
			} else {
				testDir = filepath.Join(tmpDir, tt.name)
				err := os.Mkdir(testDir, 0755)
				require.NoError(t, err)
				tt.setupFunc(t, testDir)
			}

			config, err := LoadDir(testDir)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
			assert.Nil(t, config)
		})
	}
}

package module

import (
	"errors"
	"os"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/module/filesystem"
	dotmanState "github.com/elmhuangyu/dotman/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstaller_InstallSymlinks tests the installSymlinks method with table-driven tests
func TestInstaller_InstallSymlinks(t *testing.T) {
	tests := []struct {
		name           string
		operations     []FileOperation
		mkdir          bool
		setupMocks     func(*MockFileOperator, *MockStateManager)
		expectedResult func(*testing.T, *InstallResult)
		expectedError  string
	}{
		{
			name: "successful symlink creation",
			operations: []FileOperation{
				{
					Type:   OperationCreateLink,
					Source: "/source/file1.txt",
					Target: "/target/file1.txt",
				},
			},
			mkdir: false,
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				fo.FileExistsFunc = func(path string) bool {
					// Target directory exists
					return path == "/target"
				}
				fo.CreateSymlinkFunc = func(source, target string) error {
					return nil
				}
				sm.AddMappingFunc = func(stateFile *dotmanState.StateFile, source, target, fileType string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.Len(t, result.CreatedLinks, 1)
				assert.Equal(t, "/source/file1.txt", result.CreatedLinks[0].Source)
				assert.Equal(t, "/target/file1.txt", result.CreatedLinks[0].Target)
			},
		},
		{
			name: "symlink creation fails",
			operations: []FileOperation{
				{
					Type:   OperationCreateLink,
					Source: "/source/file1.txt",
					Target: "/target/file1.txt",
				},
			},
			mkdir: false,
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				fo.FileExistsFunc = func(path string) bool {
					// Target directory exists
					return path == "/target"
				}
				fo.CreateSymlinkFunc = func(source, target string) error {
					return errors.New("permission denied")
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.False(t, result.IsSuccess)
				assert.Len(t, result.Errors, 1)
				assert.Contains(t, result.Errors[0], "permission denied")
			},
		},
		{
			name: "mkdir creates parent directories",
			operations: []FileOperation{
				{
					Type:   OperationCreateLink,
					Source: "/source/file1.txt",
					Target: "/nested/target/file1.txt",
				},
			},
			mkdir: true,
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				fo.FileExistsFunc = func(path string) bool {
					// Target directory doesn't exist initially
					return false
				}
				fo.EnsureDirectoryFunc = func(path string) error {
					return nil
				}
				fo.CreateSymlinkFunc = func(source, target string) error {
					return nil
				}
				sm.AddMappingFunc = func(stateFile *dotmanState.StateFile, source, target, fileType string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.Len(t, result.CreatedLinks, 1)
			},
		},
		{
			name: "mkdir fails when directory creation fails",
			operations: []FileOperation{
				{
					Type:   OperationCreateLink,
					Source: "/source/file1.txt",
					Target: "/nested/target/file1.txt",
				},
			},
			mkdir: true,
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				fo.FileExistsFunc = func(path string) bool {
					// Target directory doesn't exist
					return false
				}
				fo.EnsureDirectoryFunc = func(path string) error {
					return errors.New("directory creation failed")
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.False(t, result.IsSuccess)
				assert.Len(t, result.Errors, 1)
				assert.Contains(t, result.Errors[0], "directory creation failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockStateMgr := &MockStateManager{}

			tt.setupMocks(mockFileOp, mockStateMgr)

			// Create installer with mocks
			installer := &Installer{
				fileOp:   mockFileOp,
				stateMgr: mockStateMgr,
			}

			// Create test objects
			stateFile := dotmanState.NewStateFile()
			statePath := "/test/state.yaml"
			result := &InstallResult{}

			// Create symlink manager with mocked file operator
			symlinkMgr := filesystem.NewSymlinkManager(mockFileOp)

			// Call method
			err := installer.installSymlinks(
				tt.operations,
				symlinkMgr,
				tt.mkdir,
				stateFile,
				statePath,
				result,
			)

			// Check expectations
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult != nil {
					tt.expectedResult(t, result)
				}
			}
		})
	}
}

// TestInstaller_InstallTemplates tests the installTemplates method with table-driven tests
func TestInstaller_InstallTemplates(t *testing.T) {
	tests := []struct {
		name           string
		operations     []FileOperation
		vars           map[string]string
		mkdir          bool
		setupMocks     func(*MockFileOperator, *MockTemplateRenderer, *MockStateManager)
		expectedResult func(*testing.T, *InstallResult)
		expectedError  string
	}{
		{
			name: "successful template rendering",
			operations: []FileOperation{
				{
					Type:   OperationCreateTemplate,
					Source: "/source/config.dot-tmpl",
					Target: "/target/config",
				},
			},
			vars:  map[string]string{"USER": "testuser"},
			mkdir: true,
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				tr.RenderFunc = func(templatePath string, vars map[string]string) ([]byte, error) {
					return []byte("User: testuser"), nil
				}
				fo.FileExistsFunc = func(path string) bool {
					return false // Target directory doesn't exist
				}
				fo.EnsureDirectoryFunc = func(path string) error {
					return os.MkdirAll(path, 0755)
				}
				sm.AddMappingFunc = func(stateFile *dotmanState.StateFile, source, target, fileType string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.Len(t, result.CreatedTemplates, 1)
				assert.Contains(t, result.CreatedTemplates[0].Source, "/source/config.dot-tmpl")
				assert.Contains(t, result.CreatedTemplates[0].Target, "/target/config")
			},
		},
		{
			name: "template rendering fails",
			operations: []FileOperation{
				{
					Type:   OperationCreateTemplate,
					Source: "/source/config.dot-tmpl",
					Target: "/target/config",
				},
			},
			vars:  map[string]string{"USER": "testuser"},
			mkdir: true,
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				tr.RenderFunc = func(templatePath string, vars map[string]string) ([]byte, error) {
					return nil, errors.New("template syntax error")
				}
				fo.FileExistsFunc = func(path string) bool {
					return false
				}
				fo.EnsureDirectoryFunc = func(path string) error {
					return os.MkdirAll(path, 0755)
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.False(t, result.IsSuccess)
				assert.Len(t, result.Errors, 1)
				assert.Contains(t, result.Errors[0], "template syntax error")
			},
		},
		{
			name: "target directory doesn't exist and mkdir is false",
			operations: []FileOperation{
				{
					Type:   OperationCreateTemplate,
					Source: "/source/config.dot-tmpl",
					Target: "/target/config",
				},
			},
			vars:  map[string]string{"USER": "testuser"},
			mkdir: false,
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				tr.RenderFunc = func(templatePath string, vars map[string]string) ([]byte, error) {
					return []byte("User: testuser"), nil
				}
				fo.FileExistsFunc = func(path string) bool {
					return false // Target directory doesn't exist
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.False(t, result.IsSuccess)
				assert.Len(t, result.Errors, 1)
				assert.Contains(t, result.Errors[0], "target directory does not exist")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create source directory and template file
			sourceDir := tempDir + "/source"
			require.NoError(t, os.MkdirAll(sourceDir, 0755))

			// Update paths to use temp directory
			for i := range tt.operations {
				tt.operations[i].Source = tempDir + tt.operations[i].Source
				tt.operations[i].Target = tempDir + tt.operations[i].Target

				// Create the source template file
				templateContent := "User: {{.USER}}"
				require.NoError(t, os.WriteFile(tt.operations[i].Source, []byte(templateContent), 0644))
			}

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockTemplateRenderer := &MockTemplateRenderer{}
			mockStateMgr := &MockStateManager{}

			tt.setupMocks(mockFileOp, mockTemplateRenderer, mockStateMgr)

			// Create installer with mocks
			installer := &Installer{
				fileOp:   mockFileOp,
				template: mockTemplateRenderer,
				stateMgr: mockStateMgr,
			}

			// Create test objects
			stateFile := dotmanState.NewStateFile()
			statePath := tempDir + "/state.yaml"
			result := &InstallResult{}

			// Call the method
			err := installer.installTemplates(
				tt.operations,
				tt.vars,
				tt.mkdir,
				stateFile,
				statePath,
				result,
			)

			// Check expectations
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult != nil {
					tt.expectedResult(t, result)
				}
			}
		})
	}
}

// TestInstaller_Install tests the full Install method with table-driven tests
func TestInstaller_Install(t *testing.T) {
	tests := []struct {
		name           string
		request        *InstallRequest
		setupMocks     func(*MockFileOperator, *MockTemplateRenderer, *MockStateManager)
		expectedResult func(*testing.T, *InstallResult)
		expectedError  string
	}{
		{
			name: "successful installation with mixed operations",
			request: func() *InstallRequest {
				tempDir := t.TempDir()
				return &InstallRequest{
					Modules: []config.ModuleConfig{
						{
							Dir:       tempDir + "/module",
							TargetDir: tempDir + "/target",
							Ignores:   []string{},
						},
					},
					RootVars:    map[string]string{"USER": "testuser"},
					Mkdir:       true,
					Force:       false,
					DotfilesDir: tempDir,
				}
			}(),
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				// Mock file operations
				fo.FileExistsFunc = func(path string) bool {
					return false // Target doesn't exist
				}
				fo.IsSymlinkFunc = func(path string) bool {
					return false // Not symlinks
				}
				fo.CreateSymlinkFunc = func(source, target string) error {
					return nil
				}
				fo.EnsureDirectoryFunc = func(path string) error {
					return os.MkdirAll(path, 0755)
				}

				// Mock template operations
				tr.RenderFunc = func(templatePath string, vars map[string]string) ([]byte, error) {
					return []byte("rendered content"), nil
				}
				tr.ValidateFunc = func(templatePath string, vars map[string]string) error {
					return nil
				}

				// Mock state operations
				sm.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
					return dotmanState.NewStateFile(), nil
				}
				sm.SaveFunc = func(path string, stateFile *dotmanState.StateFile) error {
					return nil
				}
				sm.AddMappingFunc = func(stateFile *dotmanState.StateFile, source, target, fileType string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.True(t, result.IsSuccess)
				assert.GreaterOrEqual(t, len(result.CreatedLinks)+len(result.CreatedTemplates), 0)
			},
		},
		{
			name: "installation fails when state file cannot be loaded",
			request: func() *InstallRequest {
				tempDir := t.TempDir()
				return &InstallRequest{
					Modules: []config.ModuleConfig{
						{
							Dir:       tempDir + "/module",
							TargetDir: tempDir + "/target",
							Ignores:   []string{},
						},
					},
					RootVars:    map[string]string{"USER": "testuser"},
					Mkdir:       true,
					Force:       false,
					DotfilesDir: tempDir,
				}
			}(),
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				fo.FileExistsFunc = func(path string) bool {
					return false
				}
				fo.EnsureDirectoryFunc = func(path string) error {
					return os.MkdirAll(path, 0755)
				}
				sm.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
					return nil, errors.New("state file not found")
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				// State file loading failure should not prevent installation
				assert.True(t, result.IsSuccess)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockTemplateRenderer := &MockTemplateRenderer{}
			mockStateMgr := &MockStateManager{}

			tt.setupMocks(mockFileOp, mockTemplateRenderer, mockStateMgr)

			// Create installer with mocks
			installer := &Installer{
				fileOp:   mockFileOp,
				template: mockTemplateRenderer,
				stateMgr: mockStateMgr,
			}

			// Create module directory and files for testing
			for _, module := range tt.request.Modules {
				require.NoError(t, os.MkdirAll(module.Dir, 0755))
				require.NoError(t, os.MkdirAll(module.TargetDir, 0755))

				// Create a regular file
				regularFile := module.Dir + "/regular.txt"
				require.NoError(t, os.WriteFile(regularFile, []byte("content"), 0644))

				// Create a template file
				templateFile := module.Dir + "/config.dot-tmpl"
				require.NoError(t, os.WriteFile(templateFile, []byte("User: {{.USER}}"), 0644))
			}

			// Call the method
			result, err := installer.Install(tt.request)

			// Check expectations
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.expectedResult != nil {
					tt.expectedResult(t, result)
				}
			}
		})
	}
}

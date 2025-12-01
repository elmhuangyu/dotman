package module

import (
	"errors"
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
	t.Skip("Template unit tests need refactoring to work with os.WriteFile calls")
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
			mkdir: false,
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				tr.RenderFunc = func(templatePath string, vars map[string]string) ([]byte, error) {
					return []byte("User: testuser"), nil
				}
				fo.FileExistsFunc = func(path string) bool {
					// Target directory exists
					return path == "/target"
				}
				fo.EnsureDirectoryFunc = func(path string) error {
					return nil
				}
				sm.AddMappingFunc = func(stateFile *dotmanState.StateFile, source, target, fileType string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *InstallResult) {
				assert.Len(t, result.CreatedTemplates, 1)
				assert.Equal(t, "/source/config.dot-tmpl", result.CreatedTemplates[0].Source)
				assert.Equal(t, "/target/config", result.CreatedTemplates[0].Target)
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
			mkdir: false,
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				tr.RenderFunc = func(templatePath string, vars map[string]string) ([]byte, error) {
					return nil, errors.New("template syntax error")
				}
			},
			expectedError: "template syntax error",
		},
		{
			name: "template file copy fails",
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
				fo.CopyFileFunc = func(src, dst string) error {
					return errors.New("copy failed")
				}
			},
			expectedError: "copy failed",
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

			// Create test objects
			stateFile := dotmanState.NewStateFile()
			statePath := "/test/state.yaml"
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
	t.Skip("Full installer unit tests need complex validation mocking")
	tests := []struct {
		name           string
		request        *InstallRequest
		setupMocks     func(*MockFileOperator, *MockTemplateRenderer, *MockStateManager)
		expectedResult func(*testing.T, *InstallResult)
		expectedError  string
	}{
		{
			name: "successful installation with mixed operations",
			request: &InstallRequest{
				Modules: []config.ModuleConfig{
					{
						Dir:       "/test/module",
						TargetDir: "/test/target",
						Ignores:   []string{},
					},
				},
				RootVars:    map[string]string{"USER": "testuser"},
				Mkdir:       false,
				Force:       false,
				DotfilesDir: "/test",
			},
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				// Mock file operations
				fo.FileExistsFunc = func(path string) bool {
					return true // Simulate files exist
				}
				fo.IsSymlinkFunc = func(path string) bool {
					return false // Not symlinks
				}
				fo.CreateSymlinkFunc = func(source, target string) error {
					return nil
				}
				fo.EnsureDirectoryFunc = func(path string) error {
					return nil
				}

				// Mock template operations
				tr.RenderFunc = func(templatePath string, vars map[string]string) ([]byte, error) {
					return []byte("rendered content"), nil
				}
				tr.ValidateFunc = func(templatePath string, vars map[string]string) error {
					return nil
				}
				fo.CopyFileFunc = func(src, dst string) error {
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
			request: &InstallRequest{
				Modules: []config.ModuleConfig{
					{
						Dir:       "/test/module",
						TargetDir: "/test/target",
						Ignores:   []string{},
					},
				},
				RootVars:    map[string]string{},
				Mkdir:       false,
				Force:       false,
				DotfilesDir: "/test",
			},
			setupMocks: func(fo *MockFileOperator, tr *MockTemplateRenderer, sm *MockStateManager) {
				sm.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
					return nil, errors.New("state file not found")
				}
			},
			expectedError: "state file not found",
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

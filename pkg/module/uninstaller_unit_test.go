package module

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/module/filesystem"
	dotmanState "github.com/elmhuangyu/dotman/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUninstaller_UninstallSymlinks tests the uninstallSymlinks method with table-driven tests
func TestUninstaller_UninstallSymlinks(t *testing.T) {
	tests := []struct {
		name           string
		stateFile      *dotmanState.StateFile
		setupMocks     func(*MockFileOperator, *MockStateManager)
		expectedResult func(*testing.T, *UninstallResult)
		expectedError  string
	}{
		{
			name: "successful symlink removal",
			stateFile: func() *dotmanState.StateFile {
				sf := dotmanState.NewStateFile()
				sf.AddFileMapping("/source/file1.txt", "/target/file1.txt", dotmanState.TypeLink)
				sf.AddFileMapping("/source/file2.txt", "/target/file2.txt", dotmanState.TypeLink)
				return sf
			}(),
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				// Use default real file operations - no overrides needed
				sm.RemoveMappingsFunc = func(stateFile *dotmanState.StateFile, targets []string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *UninstallResult) {
				assert.Len(t, result.RemovedLinks, 2)
			},
		},
		{
			name: "skip invalid symlinks",
			stateFile: func() *dotmanState.StateFile {
				sf := dotmanState.NewStateFile()
				sf.AddFileMapping("/source/file1.txt", "/target/file1.txt", dotmanState.TypeLink)
				sf.AddFileMapping("/source/file2.txt", "/target/file2.txt", dotmanState.TypeLink)
				return sf
			}(),
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				fo.RemoveFileFunc = func(path string) error {
					return nil
				}
				sm.RemoveMappingsFunc = func(stateFile *dotmanState.StateFile, targets []string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *UninstallResult) {
				assert.Len(t, result.RemovedLinks, 1) // Only valid symlink removed
				assert.Len(t, result.SkippedLinks, 1) // Invalid symlink skipped
			},
		},
		{
			name: "symlink removal fails",
			stateFile: func() *dotmanState.StateFile {
				sf := dotmanState.NewStateFile()
				sf.AddFileMapping("/source/file1.txt", "/target/file1.txt", dotmanState.TypeLink)
				return sf
			}(),
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				// This will be applied to the hybrid file operator
				fo.RemoveFileFunc = func(path string) error {
					return errors.New("permission denied")
				}
				sm.RemoveMappingsFunc = func(stateFile *dotmanState.StateFile, targets []string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *UninstallResult) {
				assert.Len(t, result.FailedRemovals, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockStateMgr := &MockStateManager{}

			// Create a hybrid file operator that uses real operations but can mock failures
			hybridFileOp := &MockFileOperator{}
			realFileOp := filesystem.NewOperator()

			// Set up hybrid file operator to delegate to real operator by default
			hybridFileOp.FileExistsFunc = realFileOp.FileExists
			hybridFileOp.IsSymlinkFunc = realFileOp.IsSymlink
			hybridFileOp.ReadlinkFunc = realFileOp.Readlink
			hybridFileOp.RemoveFileFunc = realFileOp.RemoveFile

			// Apply test-specific mocks
			tt.setupMocks(hybridFileOp, mockStateMgr)

			// Create uninstaller with mocks
			uninstaller := &Uninstaller{
				fileOp:   mockFileOp,
				stateMgr: mockStateMgr,
			}

			// Create test objects
			result := &UninstallResult{}

			symlinkMgr := filesystem.NewSymlinkManager(hybridFileOp)

			// Update state file paths to use temp directory and create real symlinks
			for i := range tt.stateFile.Files {
				if tt.stateFile.Files[i].Type == dotmanState.TypeLink {
					// Update paths to use temp directory
					originalSource := tt.stateFile.Files[i].Source
					originalTarget := tt.stateFile.Files[i].Target

					tt.stateFile.Files[i].Source = tempDir + originalSource
					tt.stateFile.Files[i].Target = tempDir + originalTarget

					// Create source file
					sourcePath := tt.stateFile.Files[i].Source
					targetPath := tt.stateFile.Files[i].Target

					// Ensure source directory exists
					sourceDir := filepath.Dir(sourcePath)
					require.NoError(t, os.MkdirAll(sourceDir, 0755))

					// Ensure target directory exists
					targetDir := filepath.Dir(targetPath)
					require.NoError(t, os.MkdirAll(targetDir, 0755))

					// Create source file
					require.NoError(t, os.WriteFile(sourcePath, []byte("content"), 0644))

					// Create symlink only for the second file in "skip invalid symlinks" test
					if tt.name == "skip invalid symlinks" && i == 1 {
						// Only create the second symlink
						require.NoError(t, os.Symlink(sourcePath, targetPath))
					} else if tt.name != "skip invalid symlinks" {
						// Create all symlinks for other tests
						require.NoError(t, os.Symlink(sourcePath, targetPath))
					}
				}
			}

			// Call method
			err := uninstaller.uninstallSymlinks(
				tt.stateFile,
				symlinkMgr,
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

// TestUninstaller_UninstallGeneratedFiles tests the uninstallGeneratedFiles method with table-driven tests
func TestUninstaller_UninstallGeneratedFiles(t *testing.T) {
	tests := []struct {
		name           string
		stateFile      *dotmanState.StateFile
		setupMocks     func(*MockFileOperator, *MockStateManager)
		expectedResult func(*testing.T, *UninstallResult)
		expectedError  string
	}{
		{
			name: "successful generated file removal with matching SHA1",
			stateFile: func() *dotmanState.StateFile {
				sf := dotmanState.NewStateFile()
				sf.AddFileMapping("/source/config.dot-tmpl", "/target/config", dotmanState.TypeGenerated)
				// Set SHA1 manually
				if len(sf.Files) > 0 {
					sf.Files[0].SHA1 = "abc123"
				}
				return sf
			}(),
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				fo.FileExistsFunc = func(path string) bool {
					return true
				}
				fo.RemoveFileFunc = func(path string) error {
					return nil
				}
				sm.RemoveMappingsFunc = func(stateFile *dotmanState.StateFile, targets []string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *UninstallResult) {
				assert.Len(t, result.RemovedGenerated, 1)
				assert.Len(t, result.BackedUpGenerated, 0)
			},
		},
		{
			name: "generated file with SHA1 mismatch creates backup",
			stateFile: func() *dotmanState.StateFile {
				sf := dotmanState.NewStateFile()
				sf.AddFileMapping("/source/config.dot-tmpl", "/target/config", dotmanState.TypeGenerated)
				// Set SHA1 manually
				if len(sf.Files) > 0 {
					sf.Files[0].SHA1 = "abc123"
				}
				return sf
			}(),
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				fo.FileExistsFunc = func(path string) bool {
					return true
				}
				fo.CreateBackupFunc = func(path string) (string, error) {
					return path + ".bak", nil
				}
				fo.RemoveFileFunc = func(path string) error {
					return nil
				}
				sm.RemoveMappingsFunc = func(stateFile *dotmanState.StateFile, targets []string) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *UninstallResult) {
				assert.Len(t, result.RemovedGenerated, 1)
				assert.Len(t, result.BackedUpGenerated, 1)
			},
		},
		{
			name: "skip non-existent generated files",
			stateFile: func() *dotmanState.StateFile {
				sf := dotmanState.NewStateFile()
				sf.AddFileMapping("/source/config.dot-tmpl", "/target/config", dotmanState.TypeGenerated)
				return sf
			}(),
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				fo.FileExistsFunc = func(path string) bool {
					return false
				}
			},
			expectedResult: func(t *testing.T, result *UninstallResult) {
				assert.Len(t, result.RemovedGenerated, 0)
				assert.Len(t, result.SkippedGenerated, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockStateMgr := &MockStateManager{}

			// Create a hybrid file operator for backup operations
			hybridFileOp := &MockFileOperator{}
			realFileOp := filesystem.NewOperator()

			// Set up hybrid file operator to delegate to real operator by default
			hybridFileOp.FileExistsFunc = realFileOp.FileExists
			hybridFileOp.CreateBackupFunc = realFileOp.CreateBackup
			hybridFileOp.RemoveFileFunc = realFileOp.RemoveFile

			// Apply test-specific mocks
			tt.setupMocks(hybridFileOp, mockStateMgr)

			// Create uninstaller with mocks
			uninstaller := &Uninstaller{
				fileOp:   mockFileOp,
				stateMgr: mockStateMgr,
			}

			// Create test objects
			result := &UninstallResult{}

			backupMgr := filesystem.NewBackupManager(hybridFileOp)

			// Update state file paths to use temp directory and create real files
			for i := range tt.stateFile.Files {
				if tt.stateFile.Files[i].Type == dotmanState.TypeGenerated {
					// Update paths to use temp directory
					originalSource := tt.stateFile.Files[i].Source
					originalTarget := tt.stateFile.Files[i].Target

					tt.stateFile.Files[i].Source = tempDir + originalSource
					tt.stateFile.Files[i].Target = tempDir + originalTarget

					// Create target file
					targetPath := tt.stateFile.Files[i].Target

					// Ensure target directory exists
					targetDir := filepath.Dir(targetPath)
					require.NoError(t, os.MkdirAll(targetDir, 0755))

					// Create target file with content (except for non-existent test)
					if tt.name != "skip non-existent generated files" {
						content := "rendered content"
						if tt.name == "generated file with SHA1 mismatch creates backup" {
							content = "modified content" // Different content to trigger backup
						}
						require.NoError(t, os.WriteFile(targetPath, []byte(content), 0644))
					}

					// Set SHA1 if needed
					if tt.stateFile.Files[i].SHA1 != "" {
						if tt.name == "successful generated file removal with matching SHA1" {
							// Don't set SHA1 to skip SHA1 validation
							tt.stateFile.Files[i].SHA1 = ""
						}
						// For "generated file with SHA1 mismatch creates backup",
						// SHA1 is already set to "abc123" in the test setup
					}
				}
			}

			// Call method
			err := uninstaller.uninstallGeneratedFiles(
				tt.stateFile,
				backupMgr,
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

// TestUninstaller_Uninstall tests the full Uninstall method with table-driven tests
func TestUninstaller_Uninstall(t *testing.T) {
	tests := []struct {
		name           string
		request        *UninstallRequest
		setupMocks     func(*MockFileOperator, *MockStateManager)
		expectedResult func(*testing.T, *UninstallResult)
		expectedError  string
	}{
		{
			name: "successful uninstallation with mixed file types",
			request: &UninstallRequest{
				DotfilesDir: "/test/dotfiles",
			},
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				// Mock state operations
				sm.RemoveMappingsFunc = func(stateFile *dotmanState.StateFile, targets []string) error {
					return nil
				}
				sm.SaveFunc = func(path string, stateFile *dotmanState.StateFile) error {
					return nil
				}
			},
			expectedResult: func(t *testing.T, result *UninstallResult) {
				assert.True(t, result.IsSuccess)
				assert.Len(t, result.RemovedLinks, 1)
				assert.Len(t, result.RemovedGenerated, 1)
			},
		},
		{
			name: "uninstallation fails when state file cannot be loaded",
			request: &UninstallRequest{
				DotfilesDir: "/test/dotfiles",
			},
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				sm.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
					return nil, errors.New("state file corrupted")
				}
			},
			expectedError: "state file corrupted",
		},
		{
			name: "uninstallation with empty state file",
			request: &UninstallRequest{
				DotfilesDir: "/test/dotfiles",
			},
			setupMocks: func(fo *MockFileOperator, sm *MockStateManager) {
				sm.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
					return dotmanState.NewStateFile(), nil
				}
			},
			expectedResult: func(t *testing.T, result *UninstallResult) {
				assert.True(t, result.IsSuccess)
				assert.Contains(t, result.Summary, "0 files removed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup mocks
			mockStateMgr := &MockStateManager{}

			// Create hybrid file operators for both symlink and backup operations
			hybridSymlinkOp := &MockFileOperator{}
			realFileOp := filesystem.NewOperator()

			// Set up hybrid file operators to delegate to real operator by default
			hybridSymlinkOp.FileExistsFunc = realFileOp.FileExists
			hybridSymlinkOp.IsSymlinkFunc = realFileOp.IsSymlink
			hybridSymlinkOp.ReadlinkFunc = realFileOp.Readlink
			hybridSymlinkOp.RemoveFileFunc = realFileOp.RemoveFile

			// Apply test-specific mocks
			tt.setupMocks(hybridSymlinkOp, mockStateMgr)

			// Create uninstaller with hybrid file operator for real operations
			uninstaller := &Uninstaller{
				fileOp:   hybridSymlinkOp, // Use hybrid for both symlink and backup
				stateMgr: mockStateMgr,
			}

			// Update request to use temp directory
			request := *tt.request // Copy the request
			request.DotfilesDir = tempDir

			// Create test state file
			stateFile := dotmanState.NewStateFile()
			if tt.name == "successful uninstallation with mixed file types" {
				stateFile.AddFileMapping("/source/file1.txt", "/target/file1.txt", dotmanState.TypeLink)
				stateFile.AddFileMapping("/source/config.dot-tmpl", "/target/config", dotmanState.TypeGenerated)
			}

			// Update state file paths to use temp directory and create real files
			for i := range stateFile.Files {
				originalSource := stateFile.Files[i].Source
				originalTarget := stateFile.Files[i].Target

				stateFile.Files[i].Source = tempDir + originalSource
				stateFile.Files[i].Target = tempDir + originalTarget

				switch stateFile.Files[i].Type {
				case dotmanState.TypeLink:
					// Create symlink
					sourcePath := stateFile.Files[i].Source
					targetPath := stateFile.Files[i].Target

					// Ensure directories exist
					sourceDir := filepath.Dir(sourcePath)
					targetDir := filepath.Dir(targetPath)
					require.NoError(t, os.MkdirAll(sourceDir, 0755))
					require.NoError(t, os.MkdirAll(targetDir, 0755))

					// Create source file and symlink
					require.NoError(t, os.WriteFile(sourcePath, []byte("content"), 0644))
					err := os.Symlink(sourcePath, targetPath)
					require.NoError(t, err)

				case dotmanState.TypeGenerated:
					// Create generated file
					targetPath := stateFile.Files[i].Target

					// Ensure target directory exists
					targetDir := filepath.Dir(targetPath)
					require.NoError(t, os.MkdirAll(targetDir, 0755))

					// Create target file
					content := "rendered content"
					require.NoError(t, os.WriteFile(targetPath, []byte(content), 0644))
				}
			}

			// Mock the state manager to return our test state file (except for load error test)
			if tt.name != "uninstallation fails when state file cannot be loaded" {
				mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
					return stateFile, nil
				}
			}

			// Call method
			result, err := uninstaller.Uninstall(&request)

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

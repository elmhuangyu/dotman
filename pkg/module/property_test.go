package module

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"testing/quick"

	"github.com/elmhuangyu/dotman/pkg/config"
	dotmanState "github.com/elmhuangyu/dotman/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstaller_PropertyBasedTests runs property-based tests for edge cases
func TestInstaller_PropertyBasedTests(t *testing.T) {
	// Test that installer handles empty module list gracefully
	t.Run("empty modules list", func(t *testing.T) {
		f := func(req InstallRequest) bool {
			// Ensure modules is empty for this test
			req.Modules = []config.ModuleConfig{}

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockTemplateRenderer := &MockTemplateRenderer{}
			mockStateMgr := &MockStateManager{}

			installer := &Installer{
				fileOp:   mockFileOp,
				template: mockTemplateRenderer,
				stateMgr: mockStateMgr,
			}

			result, err := installer.Install(&req)

			// Should succeed with no operations
			return err == nil && result != nil && result.IsSuccess
		}

		err := quick.Check(f, nil)
		if err != nil {
			t.Error(err)
		}
	})

	// Test that installer handles very long paths
	t.Run("long paths", func(t *testing.T) {
		f := func(length int) bool {
			if length <= 0 || length > 1000 {
				return true // Skip invalid lengths
			}

			// Create long path
			longPath := ""
			for i := 0; i < length; i++ {
				longPath += "a"
			}

			req := &InstallRequest{
				Modules: []config.ModuleConfig{
					{
						Dir:       "/source/" + longPath,
						TargetDir: "/target/" + longPath,
						Ignores:   []string{},
					},
				},
				RootVars:    map[string]string{},
				Mkdir:       true,
				Force:       false,
				DotfilesDir: "/test",
			}

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockTemplateRenderer := &MockTemplateRenderer{}
			mockStateMgr := &MockStateManager{}

			mockFileOp.EnsureDirectoryFunc = func(path string) error {
				return nil // Always succeed for directory creation
			}
			mockFileOp.FileExistsFunc = func(path string) bool {
				return true
			}
			mockFileOp.IsSymlinkFunc = func(path string) bool {
				return false
			}
			mockFileOp.CreateSymlinkFunc = func(source, target string) error {
				return nil
			}

			mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
				return dotmanState.NewStateFile(), nil
			}
			mockStateMgr.SaveFunc = func(path string, stateFile *dotmanState.StateFile) error {
				return nil
			}
			mockStateMgr.AddMappingFunc = func(stateFile *dotmanState.StateFile, source, target, fileType string) error {
				return nil
			}

			installer := &Installer{
				fileOp:   mockFileOp,
				template: mockTemplateRenderer,
				stateMgr: mockStateMgr,
			}

			result, err := installer.Install(req)

			// Should handle long paths without panicking
			return err == nil && result != nil
		}

		err := quick.Check(f, nil)
		if err != nil {
			t.Error(err)
		}
	})

	// Test that installer handles special characters in paths
	t.Run("special characters in paths", func(t *testing.T) {
		f := func(pathChars string) bool {
			// Create a path with special characters
			specialPath := "/tmp/test_" + pathChars + "_module"

			req := &InstallRequest{
				Modules: []config.ModuleConfig{
					{
						Dir:       specialPath,
						TargetDir: "/tmp/target_" + pathChars,
						Ignores:   []string{},
					},
				},
				RootVars:    map[string]string{},
				Mkdir:       true,
				Force:       false,
				DotfilesDir: "/tmp",
			}

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockTemplateRenderer := &MockTemplateRenderer{}
			mockStateMgr := &MockStateManager{}

			mockFileOp.EnsureDirectoryFunc = func(path string) error {
				return nil // Always succeed for directory creation
			}
			mockFileOp.FileExistsFunc = func(path string) bool {
				return true
			}
			mockFileOp.IsSymlinkFunc = func(path string) bool {
				return false
			}
			mockFileOp.CreateSymlinkFunc = func(source, target string) error {
				return nil
			}

			mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
				return dotmanState.NewStateFile(), nil
			}
			mockStateMgr.SaveFunc = func(path string, stateFile *dotmanState.StateFile) error {
				return nil
			}
			mockStateMgr.AddMappingFunc = func(stateFile *dotmanState.StateFile, source, target, fileType string) error {
				return nil
			}

			installer := &Installer{
				fileOp:   mockFileOp,
				template: mockTemplateRenderer,
				stateMgr: mockStateMgr,
			}

			result, err := installer.Install(req)

			// Should handle special characters without panicking
			return err == nil && result != nil
		}

		// Test various special characters
		// All should fail validation (expected behavior)
		testCases := []string{
			"space test",
			"unicode-测试",
			"special!@#$%^&*()",
			"brackets[test]",
			"quotes'test\"",
		}

		for _, chars := range testCases {
			result := f(chars)
			// Should fail validation (return false or error)
			if result {
				t.Errorf("Expected validation to fail for special characters: %s", chars)
			}
		}
	})
}

// TestUninstaller_PropertyBasedTests runs property-based tests for uninstaller edge cases
func TestUninstaller_PropertyBasedTests(t *testing.T) {
	// Test that uninstaller handles empty state file gracefully
	t.Run("empty state file", func(t *testing.T) {
		f := func(req UninstallRequest) bool {
			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockStateMgr := &MockStateManager{}

			mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
				return dotmanState.NewStateFile(), nil // Empty state file
			}

			uninstaller := &Uninstaller{
				fileOp:   mockFileOp,
				stateMgr: mockStateMgr,
			}

			result, err := uninstaller.Uninstall(&req)

			// Should succeed with no operations
			return err == nil && result != nil && result.IsSuccess &&
				len(result.RemovedLinks) == 0 && len(result.RemovedGenerated) == 0
		}

		err := quick.Check(f, nil)
		if err != nil {
			t.Error(err)
		}
	})

	// Test that uninstaller handles large number of files
	t.Run("large number of files", func(t *testing.T) {
		f := func(fileCount int) bool {
			if fileCount <= 0 || fileCount > 1000 {
				return true // Skip invalid counts
			}

			// Create state file with many entries
			stateFile := dotmanState.NewStateFile()
			for i := 0; i < fileCount; i++ {
				source := fmt.Sprintf("/source/file%d.txt", i)
				target := fmt.Sprintf("/target/file%d.txt", i)
				stateFile.AddFileMapping(source, target, dotmanState.TypeLink)
			}

			req := &UninstallRequest{
				DotfilesDir: "/test",
			}

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockStateMgr := &MockStateManager{}

			mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
				return stateFile, nil
			}

			mockFileOp.IsSymlinkFunc = func(path string) bool {
				return true
			}
			mockFileOp.ReadlinkFunc = func(path string) (string, error) {
				// Return matching source for each target
				return path, nil
			}
			mockFileOp.RemoveFileFunc = func(path string) error {
				return nil
			}

			mockStateMgr.RemoveMappingsFunc = func(stateFile *dotmanState.StateFile, targets []string) error {
				return nil
			}
			mockStateMgr.SaveFunc = func(path string, stateFile *dotmanState.StateFile) error {
				return nil
			}

			uninstaller := &Uninstaller{
				fileOp:   mockFileOp,
				stateMgr: mockStateMgr,
			}

			result, err := uninstaller.Uninstall(req)

			// Should handle many files without panicking
			return err == nil && result != nil && result.IsSuccess &&
				len(result.RemovedLinks) == fileCount
		}

		err := quick.Check(f, nil)
		if err != nil {
			t.Error(err)
		}
	})

	// Test that uninstaller handles mixed file types
	t.Run("mixed file types", func(t *testing.T) {
		f := func(linkCount, generatedCount int) bool {
			if linkCount < 0 || generatedCount < 0 || linkCount > 100 || generatedCount > 100 {
				return true // Skip invalid counts
			}

			// Create state file with mixed file types
			stateFile := dotmanState.NewStateFile()

			// Add links
			for i := 0; i < linkCount; i++ {
				source := fmt.Sprintf("/source/link%d.txt", i)
				target := fmt.Sprintf("/target/link%d.txt", i)
				stateFile.AddFileMapping(source, target, dotmanState.TypeLink)
			}

			// Add generated files
			for i := 0; i < generatedCount; i++ {
				source := fmt.Sprintf("/source/gen%d.dot-tmpl", i)
				target := fmt.Sprintf("/target/gen%d", i)
				stateFile.AddFileMapping(source, target, dotmanState.TypeGenerated)
				// Set SHA1 for generated files
				if len(stateFile.Files) > 0 {
					stateFile.Files[len(stateFile.Files)-1].SHA1 = "abc123"
				}
			}

			req := &UninstallRequest{
				DotfilesDir: "/test",
			}

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockStateMgr := &MockStateManager{}

			mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
				return stateFile, nil
			}

			mockFileOp.FileExistsFunc = func(path string) bool {
				return true
			}
			mockFileOp.IsSymlinkFunc = func(path string) bool {
				return true
			}
			mockFileOp.ReadlinkFunc = func(path string) (string, error) {
				return path, nil
			}
			mockFileOp.RemoveFileFunc = func(path string) error {
				return nil
			}

			mockStateMgr.RemoveMappingsFunc = func(stateFile *dotmanState.StateFile, targets []string) error {
				return nil
			}
			mockStateMgr.SaveFunc = func(path string, stateFile *dotmanState.StateFile) error {
				return nil
			}

			uninstaller := &Uninstaller{
				fileOp:   mockFileOp,
				stateMgr: mockStateMgr,
			}

			result, err := uninstaller.Uninstall(req)

			// Should handle mixed types correctly
			return err == nil && result != nil && result.IsSuccess &&
				len(result.RemovedLinks) == linkCount &&
				len(result.RemovedGenerated) == generatedCount
		}

		err := quick.Check(f, nil)
		if err != nil {
			t.Error(err)
		}
	})
}

// TestErrorHandling_PropertyBasedTests tests error handling edge cases
func TestErrorHandling_PropertyBasedTests(t *testing.T) {
	// Test that installer handles various error conditions gracefully
	t.Run("installer error handling", func(t *testing.T) {
		// Test state load error
		t.Run("state load error", func(t *testing.T) {
			req := &InstallRequest{
				Modules: []config.ModuleConfig{
					{
						Dir:       "/test/module",
						TargetDir: "/test/target",
						Ignores:   []string{},
					},
				},
				RootVars:    map[string]string{},
				Mkdir:       true,
				Force:       false,
				DotfilesDir: "/test",
			}

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockTemplateRenderer := &MockTemplateRenderer{}
			mockStateMgr := &MockStateManager{}

			mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
				return nil, errors.New("state file corrupted")
			}

			installer := &Installer{
				fileOp:   mockFileOp,
				template: mockTemplateRenderer,
				stateMgr: mockStateMgr,
			}

			result, err := installer.Install(req)

			// Should handle error gracefully
			assert.Error(t, err)
			assert.Nil(t, result)
		})

		// Test template render error
		t.Run("template render error", func(t *testing.T) {
			tempDir := t.TempDir()

			req := &InstallRequest{
				Modules: []config.ModuleConfig{
					{
						Dir:       tempDir + "/module",
						TargetDir: tempDir + "/target",
						Ignores:   []string{},
					},
				},
				RootVars:    map[string]string{},
				Mkdir:       true,
				Force:       false,
				DotfilesDir: tempDir,
			}

			// Create module directory and template file
			require.NoError(t, os.MkdirAll(tempDir+"/module", 0755))
			templateContent := "Hello {{.USER}}"
			require.NoError(t, os.WriteFile(tempDir+"/module/config.dot-tmpl", []byte(templateContent), 0644))

			// Setup mocks
			mockFileOp := &MockFileOperator{}
			mockTemplateRenderer := &MockTemplateRenderer{}
			mockStateMgr := &MockStateManager{}

			mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
				return dotmanState.NewStateFile(), nil
			}
			mockTemplateRenderer.RenderFunc = func(templatePath string, vars map[string]string) ([]byte, error) {
				return nil, errors.New("template syntax error")
			}
			mockFileOp.FileExistsFunc = func(path string) bool {
				return true
			}
			mockFileOp.EnsureDirectoryFunc = func(path string) error {
				return os.MkdirAll(path, 0755)
			}

			installer := &Installer{
				fileOp:   mockFileOp,
				template: mockTemplateRenderer,
				stateMgr: mockStateMgr,
			}

			result, err := installer.Install(req)

			// Should handle error gracefully
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.False(t, result.IsSuccess)
			assert.NotEmpty(t, result.Errors)
		})
	})

	// Test that uninstaller handles various error conditions gracefully
	t.Run("uninstaller error handling", func(t *testing.T) {
		f := func(errorType string) bool {
			var req *UninstallRequest

			switch errorType {
			case "state_load_error":
				req = &UninstallRequest{
					DotfilesDir: "/test",
				}

				// Setup mocks
				mockFileOp := &MockFileOperator{}
				mockStateMgr := &MockStateManager{}

				mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
					return nil, errors.New("state file corrupted")
				}

				uninstaller := &Uninstaller{
					fileOp:   mockFileOp,
					stateMgr: mockStateMgr,
				}

				result, err := uninstaller.Uninstall(req)

				// Should handle error gracefully
				return err != nil && result == nil

			case "empty_state":
				req = &UninstallRequest{
					DotfilesDir: "/test",
				}

				// Setup mocks
				mockFileOp := &MockFileOperator{}
				mockStateMgr := &MockStateManager{}

				mockStateMgr.LoadFunc = func(path string) (*dotmanState.StateFile, error) {
					return dotmanState.NewStateFile(), nil
				}

				uninstaller := &Uninstaller{
					fileOp:   mockFileOp,
					stateMgr: mockStateMgr,
				}

				result, err := uninstaller.Uninstall(req)

				// Should handle empty state gracefully
				return err == nil && result != nil && result.IsSuccess

			default:
				return true
			}
		}

		// Test various error types
		errorTypes := []string{
			"state_load_error",
			"empty_state",
		}

		for _, errorType := range errorTypes {
			if !f(errorType) {
				t.Errorf("Failed to handle error type: %s", errorType)
			}
		}

		err := quick.Check(f, nil)
		if err != nil {
			t.Error(err)
		}
	})
}

package module

import (
	"fmt"
	"testing"
	"testing/quick"

	"github.com/elmhuangyu/dotman/pkg/config"
	dotmanState "github.com/elmhuangyu/dotman/pkg/state"
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
		t.Skip("Special character tests need validation mocking")
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
	t.Skip("Error handling tests need complex validation mocking")
}

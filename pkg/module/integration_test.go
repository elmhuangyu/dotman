package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallUninstallCycle(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	moduleDir := filepath.Join(dotfilesDir, "module")
	targetDir := filepath.Join(tempDir, "target")

	// Create directories
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create test files in module
	sourceFile1 := filepath.Join(moduleDir, "file1.txt")
	sourceFile2 := filepath.Join(moduleDir, "file2.txt")
	sourceFile3 := filepath.Join(moduleDir, "file3.txt")

	err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(sourceFile3, []byte("content3"), 0644)
	require.NoError(t, err)

	// Create module config
	modules := []config.ModuleConfig{
		{
			Dir:       moduleDir,
			TargetDir: targetDir,
			Ignores:   []string{},
		},
	}

	t.Run("full install and uninstall cycle", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		// Step 1: Install files
		installResult, err := Install(modules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, installResult.IsSuccess)
		assert.Len(t, installResult.CreatedLinks, 3)
		assert.Len(t, installResult.SkippedLinks, 0)

		// Verify symlinks were created
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")
		targetFile3 := filepath.Join(targetDir, "file3.txt")

		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)
		assert.FileExists(t, targetFile3)

		// Verify state file was created
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile)
		assert.Len(t, stateFile.Files, 3)

		// Step 2: Uninstall files
		uninstallResult, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, uninstallResult.IsSuccess)
		assert.Len(t, uninstallResult.RemovedLinks, 3)
		assert.Len(t, uninstallResult.SkippedLinks, 0)

		// Verify symlinks were removed
		assert.NoFileExists(t, targetFile1)
		assert.NoFileExists(t, targetFile2)
		assert.NoFileExists(t, targetFile3)

		// Verify state file is empty
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 0)
	})

	t.Run("install with existing symlinks then uninstall", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		// Pre-create correct symlinks
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")
		targetFile3 := filepath.Join(targetDir, "file3.txt")

		absSource1, err := filepath.Abs(sourceFile1)
		require.NoError(t, err)

		absSource2, err := filepath.Abs(sourceFile2)
		require.NoError(t, err)

		absSource3, err := filepath.Abs(sourceFile3)
		require.NoError(t, err)

		err = os.Symlink(absSource1, targetFile1)
		require.NoError(t, err)

		err = os.Symlink(absSource2, targetFile2)
		require.NoError(t, err)

		err = os.Symlink(absSource3, targetFile3)
		require.NoError(t, err)

		// Step 1: Install (should skip existing symlinks)
		installResult, err := Install(modules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, installResult.IsSuccess)
		assert.Len(t, installResult.CreatedLinks, 0)
		assert.Len(t, installResult.SkippedLinks, 3)

		// Verify state file contains skipped files
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile)
		assert.Len(t, stateFile.Files, 3)

		// Step 2: Uninstall (should remove all symlinks)
		uninstallResult, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, uninstallResult.IsSuccess)
		assert.Len(t, uninstallResult.RemovedLinks, 3)
		assert.Len(t, uninstallResult.SkippedLinks, 0)

		// Verify symlinks were removed
		assert.NoFileExists(t, targetFile1)
		assert.NoFileExists(t, targetFile2)
		assert.NoFileExists(t, targetFile3)

		// Verify state file is empty
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 0)
	})

	t.Run("install with force mode then uninstall", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		// Create conflicting files
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")
		targetFile3 := filepath.Join(targetDir, "file3.txt")

		err = os.WriteFile(targetFile1, []byte("existing1"), 0644)
		require.NoError(t, err)

		err = os.WriteFile(targetFile2, []byte("existing2"), 0644)
		require.NoError(t, err)

		err = os.WriteFile(targetFile3, []byte("existing3"), 0644)
		require.NoError(t, err)

		// Step 1: Install with force mode
		installResult, err := Install(modules, map[string]string{}, false, true, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, installResult.IsSuccess)
		assert.Len(t, installResult.CreatedLinks, 3)
		assert.Len(t, installResult.SkippedLinks, 0)

		// Verify backup files exist
		backupFile1 := targetFile1 + ".bak"
		backupFile2 := targetFile2 + ".bak"
		backupFile3 := targetFile3 + ".bak"

		assert.FileExists(t, backupFile1)
		assert.FileExists(t, backupFile2)
		assert.FileExists(t, backupFile3)

		// Verify state file contains created symlinks (should be 3 including wrong.txt)
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile)
		assert.Len(t, stateFile.Files, 3)

		// Step 2: Uninstall (should remove symlinks but leave backups)
		uninstallResult, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, uninstallResult.IsSuccess)
		assert.Len(t, uninstallResult.RemovedLinks, 3)
		assert.Len(t, uninstallResult.SkippedLinks, 0)

		// Verify symlinks were removed but backups remain
		assert.NoFileExists(t, targetFile1)
		assert.NoFileExists(t, targetFile2)
		assert.NoFileExists(t, targetFile3)

		assert.FileExists(t, backupFile1)
		assert.FileExists(t, backupFile2)
		assert.FileExists(t, backupFile3)

		// Verify state file is empty
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 0)
	})
}

func TestUninstallWithModifiedSymlinks(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	moduleDir := filepath.Join(dotfilesDir, "module")
	targetDir := filepath.Join(tempDir, "target")

	// Create directories
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create test files
	sourceFile1 := filepath.Join(moduleDir, "file1.txt")
	sourceFile2 := filepath.Join(moduleDir, "file2.txt")
	wrongSource := filepath.Join(moduleDir, "wrong.txt")

	err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(wrongSource, []byte("wrong"), 0644)
	require.NoError(t, err)

	t.Run("uninstall skips modified symlinks with warnings", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		// Create module config that excludes wrong.txt
		testModules := []config.ModuleConfig{
			{
				Dir:       moduleDir,
				TargetDir: targetDir,
				Ignores:   []string{"wrong"},
			},
		}

		// Install files first
		installResult, err := Install(testModules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, installResult.IsSuccess)

		// Modify one symlink to point to wrong source
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")

		// Remove correct symlink and create wrong one
		os.Remove(targetFile1)
		absWrongSource, err := filepath.Abs(wrongSource)
		require.NoError(t, err)
		err = os.Symlink(absWrongSource, targetFile1)
		require.NoError(t, err)

		// Replace another symlink with a regular file
		os.Remove(targetFile2)
		err = os.WriteFile(targetFile2, []byte("regular file"), 0644)
		require.NoError(t, err)

		// Run uninstall
		uninstallResult, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, uninstallResult.IsSuccess)
		assert.Len(t, uninstallResult.RemovedLinks, 0) // None should be removed
		assert.Len(t, uninstallResult.SkippedLinks, 2) // Both should be skipped

		// Verify skipped reasons
		skipReasons := make(map[string]string)
		for _, skipped := range uninstallResult.SkippedLinks {
			skipReasons[skipped.Operation.Target] = skipped.Reason
		}

		assert.Contains(t, skipReasons[targetFile1], "symlink points to")
		assert.Contains(t, skipReasons[targetFile2], "target exists but is not a symlink")

		// Verify files still exist (unchanged)
		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)

		// Verify state file still contains both entries
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile)
		assert.Len(t, stateFile.Files, 2)
	})
}

func TestMultipleInstallUninstallCycles(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	moduleDir := filepath.Join(dotfilesDir, "module")
	targetDir := filepath.Join(tempDir, "target")

	// Create directories
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create test files
	sourceFile1 := filepath.Join(moduleDir, "file1.txt")
	sourceFile2 := filepath.Join(moduleDir, "file2.txt")

	err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
	require.NoError(t, err)

	// Create module config
	modules := []config.ModuleConfig{
		{
			Dir:       moduleDir,
			TargetDir: targetDir,
			Ignores:   []string{},
		},
	}

	t.Run("multiple install/uninstall cycles work correctly", func(t *testing.T) {
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")

		// Cycle 1: Install and uninstall
		installResult1, err := Install(modules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, installResult1.IsSuccess)
		assert.Len(t, installResult1.CreatedLinks, 2)

		uninstallResult1, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, uninstallResult1.IsSuccess)
		assert.Len(t, uninstallResult1.RemovedLinks, 2)

		// Verify state file is empty
		stateFile1, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile1)
		assert.Len(t, stateFile1.Files, 0)

		// Cycle 2: Install and uninstall again
		installResult2, err := Install(modules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, installResult2.IsSuccess)
		assert.Len(t, installResult2.CreatedLinks, 2)

		uninstallResult2, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, uninstallResult2.IsSuccess)
		assert.Len(t, uninstallResult2.RemovedLinks, 2)

		// Verify state file is empty again
		stateFile2, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile2)
		assert.Len(t, stateFile2.Files, 0)

		// Cycle 3: Install with existing symlinks (should skip)
		// Pre-create symlinks
		absSource1, err := filepath.Abs(sourceFile1)
		require.NoError(t, err)
		absSource2, err := filepath.Abs(sourceFile2)
		require.NoError(t, err)

		err = os.Symlink(absSource1, targetFile1)
		require.NoError(t, err)
		err = os.Symlink(absSource2, targetFile2)
		require.NoError(t, err)

		installResult3, err := Install(modules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, installResult3.IsSuccess)
		assert.Len(t, installResult3.CreatedLinks, 0)
		assert.Len(t, installResult3.SkippedLinks, 2)

		uninstallResult3, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, uninstallResult3.IsSuccess)
		assert.Len(t, uninstallResult3.RemovedLinks, 2)

		// Verify final state
		finalStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, finalStateFile)
		assert.Len(t, finalStateFile.Files, 0)

		assert.NoFileExists(t, targetFile1)
		assert.NoFileExists(t, targetFile2)
	})
}

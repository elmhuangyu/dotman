package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallUninstallIntegration(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	targetDir := filepath.Join(tempDir, "target")

	// Create directories
	err := os.MkdirAll(dotfilesDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create test module structure
	moduleDir := filepath.Join(dotfilesDir, "module")
	err = os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	// Create test files in module
	sourceFile1 := filepath.Join(moduleDir, "file1.txt")
	sourceFile2 := filepath.Join(moduleDir, "file2.txt")

	err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
	require.NoError(t, err)

	// Create module config file in module directory
	configPath := filepath.Join(moduleDir, "Dotfile")
	data := []byte(`target_dir: "` + targetDir + `"
ignores: []`)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	t.Run("install calls uninstall when not in dry-run mode", func(t *testing.T) {
		// Clean up any existing state
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		os.Remove(statePath)

		// First, create an existing installation by running install once
		err := install(dotfilesDir, false, false, true)
		require.NoError(t, err)

		// Verify that symlinks were created
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")

		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)

		// Verify state file was created
		_, err = os.Stat(statePath)
		assert.NoError(t, err)

		// Now run install again - this should call uninstall first
		err = install(dotfilesDir, false, false, true)
		require.NoError(t, err)

		// Verify that symlinks still exist (recreated after uninstall)
		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)

		// Verify state file still exists
		_, err = os.Stat(statePath)
		assert.NoError(t, err)
	})

	t.Run("install skips uninstall in dry-run mode", func(t *testing.T) {
		// Clean up any existing state
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		os.Remove(statePath)

		// Create an initial installation
		err := install(dotfilesDir, false, false, true)
		require.NoError(t, err)

		// Verify state file exists
		_, err = os.Stat(statePath)
		assert.NoError(t, err)

		// Run install in dry-run mode - should not call uninstall
		err = install(dotfilesDir, true, false, false)
		require.NoError(t, err)

		// State file should still exist (uninstall was not called)
		_, err = os.Stat(statePath)
		assert.NoError(t, err)
	})

	t.Run("install handles uninstall errors gracefully", func(t *testing.T) {
		// Clean up any existing state
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		os.Remove(statePath)

		// Create a corrupted state file that will cause uninstall to fail
		corruptedState := "invalid: yaml: content"
		err = os.WriteFile(statePath, []byte(corruptedState), 0644)
		require.NoError(t, err)

		// Run install - should handle uninstall error gracefully and proceed
		err = install(dotfilesDir, false, false, true)
		require.NoError(t, err)

		// Verify that installation still succeeded
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")

		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)
	})

	t.Run("install with no previous installation", func(t *testing.T) {
		// Clean up any existing state
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		os.Remove(statePath)

		// Remove any existing symlinks
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")
		os.Remove(targetFile1)
		os.Remove(targetFile2)

		// Run install with no previous installation
		err := install(dotfilesDir, false, false, true)
		require.NoError(t, err)

		// Verify that installation succeeded
		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)
	})
}

func TestInstallWithMissingStateFile(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	targetDir := filepath.Join(tempDir, "target")

	// Create directories
	err := os.MkdirAll(dotfilesDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create test module structure
	moduleDir := filepath.Join(dotfilesDir, "module")
	err = os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	// Create test files in module
	sourceFile := filepath.Join(moduleDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Create module config file in module directory
	configPath := filepath.Join(moduleDir, "Dotfile")
	data := []byte(`target_dir: "` + targetDir + `"
ignores: []`)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Ensure state file does not exist
	statePath := filepath.Join(dotfilesDir, "state.yaml")
	_, err = os.Stat(statePath)
	assert.True(t, os.IsNotExist(err))

	// Run install - should handle missing state file gracefully
	err = install(dotfilesDir, false, false, true)
	require.NoError(t, err)

	// Verify that installation succeeded
	targetFile := filepath.Join(targetDir, "file.txt")
	assert.FileExists(t, targetFile)
}

func TestInstallIntegrationFlags(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	targetDir := filepath.Join(tempDir, "target")

	// Create directories
	err := os.MkdirAll(dotfilesDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create test module structure
	moduleDir := filepath.Join(dotfilesDir, "module")
	err = os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	// Create test files in module
	sourceFile := filepath.Join(moduleDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Create module config file in module directory
	configPath := filepath.Join(moduleDir, "Dotfile")
	data := []byte(`target_dir: "` + targetDir + `"
ignores: []`)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	t.Run("install with force flag", func(t *testing.T) {
		// Clean up any existing state
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		os.Remove(statePath)

		// Create a conflicting file at target location
		targetFile := filepath.Join(targetDir, "file.txt")
		err = os.WriteFile(targetFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		// Run install with force flag - should handle uninstall first then force install
		err = install(dotfilesDir, false, true, true)
		require.NoError(t, err)

		// Verify that symlink was created (overwriting the existing file)
		assert.FileExists(t, targetFile)

		// Verify it's a symlink
		info, err := os.Lstat(targetFile)
		require.NoError(t, err)
		assert.True(t, info.Mode()&os.ModeSymlink != 0)
	})

	t.Run("install with mkdir flag", func(t *testing.T) {
		// Clean up any existing state and target
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		os.Remove(statePath)
		os.RemoveAll(targetDir)

		// Run install with mkdir flag - should create target directory
		err = install(dotfilesDir, false, false, true)
		require.NoError(t, err)

		// Verify that target directory was created and symlink exists
		targetFile := filepath.Join(targetDir, "file.txt")
		assert.FileExists(t, targetFile)

		// Verify it's a symlink
		info, err := os.Lstat(targetFile)
		require.NoError(t, err)
		assert.True(t, info.Mode()&os.ModeSymlink != 0)
	})

	t.Run("install with conflicting files after uninstall cleanup", func(t *testing.T) {
		// Clean up any existing state
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		os.Remove(statePath)

		// First installation
		err = install(dotfilesDir, false, false, true)
		require.NoError(t, err)

		// Verify first installation
		targetFile := filepath.Join(targetDir, "file.txt")
		assert.FileExists(t, targetFile)

		// Manually remove the symlink and create a conflicting file
		err = os.Remove(targetFile)
		require.NoError(t, err)
		err = os.WriteFile(targetFile, []byte("conflicting content"), 0644)
		require.NoError(t, err)

		// Run install again with force flag - should call uninstall first (which will skip the conflicting file)
		// then install will handle the conflict with force flag
		err = install(dotfilesDir, false, true, true)
		require.NoError(t, err)

		// Verify that symlink was recreated
		assert.FileExists(t, targetFile)
		info, err := os.Lstat(targetFile)
		require.NoError(t, err)
		assert.True(t, info.Mode()&os.ModeSymlink != 0)
	})
}

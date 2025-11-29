package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall(t *testing.T) {
	tempDir := t.TempDir()

	// Create test module structure
	moduleDir := filepath.Join(tempDir, "module")
	targetDir := filepath.Join(tempDir, "target")

	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create test files in module
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

	t.Run("successful installation", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		result, err := Install(modules, false, false)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 2)
		assert.Len(t, result.SkippedLinks, 0)
		assert.Len(t, result.Errors, 0)
		assert.Contains(t, result.Summary, "2 symlinks created")

		// Verify symlinks were created
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")

		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)

		// Check they are symlinks pointing to correct files using absolute paths
		linkTarget1, err := os.Readlink(targetFile1)
		require.NoError(t, err)
		absSource1, err := filepath.Abs(sourceFile1)
		require.NoError(t, err)
		assert.Equal(t, absSource1, linkTarget1)

		linkTarget2, err := os.Readlink(targetFile2)
		require.NoError(t, err)
		absSource2, err := filepath.Abs(sourceFile2)
		require.NoError(t, err)
		assert.Equal(t, absSource2, linkTarget2)
	})

	t.Run("skip existing correct symlinks", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		// Pre-create correct symlinks
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")

		err := os.Symlink(sourceFile1, targetFile1)
		require.NoError(t, err)

		err = os.Symlink(sourceFile2, targetFile2)
		require.NoError(t, err)

		result, err := Install(modules, false, false)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 0)
		assert.Len(t, result.SkippedLinks, 2)
		assert.Len(t, result.Errors, 0)
		assert.Contains(t, result.Summary, "0 symlinks created")
		assert.Contains(t, result.Summary, "2 skipped")
	})

	t.Run("fail on conflicts", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		// Create a regular file where symlink should go
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		err := os.WriteFile(targetFile1, []byte("existing file"), 0644)
		require.NoError(t, err)

		result, err := Install(modules, false, false)
		require.NoError(t, err)
		assert.False(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 0)
		assert.Len(t, result.SkippedLinks, 0)
		assert.Len(t, result.Errors, 1)
		assert.Contains(t, result.Summary, "Installation failed")
	})

	t.Run("fail when target directory does not exist", func(t *testing.T) {
		// Use nested target directory that doesn't exist
		nestedTargetDir := filepath.Join(tempDir, "nested", "target")
		nestedModules := []config.ModuleConfig{
			{
				Dir:       moduleDir,
				TargetDir: nestedTargetDir,
				Ignores:   []string{},
			},
		}

		result, err := Install(nestedModules, false, false)
		require.NoError(t, err)
		assert.False(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 0)
		assert.Len(t, result.Errors, 1) // Stops on first error
		assert.Contains(t, result.Summary, "Installation failed")
	})

	t.Run("succeed when target directory does not exist and mkdir is true", func(t *testing.T) {
		// Use nested target directory that doesn't exist
		mkdirTargetDir := filepath.Join(tempDir, "mkdir", "nested", "target")
		mkdirModules := []config.ModuleConfig{
			{
				Dir:       moduleDir,
				TargetDir: mkdirTargetDir,
				Ignores:   []string{},
			},
		}

		result, err := Install(mkdirModules, true, false)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 2)
		assert.Len(t, result.SkippedLinks, 0)
		assert.Len(t, result.Errors, 0)
		assert.Contains(t, result.Summary, "2 symlinks created")

		// Verify directories were created
		assert.DirExists(t, mkdirTargetDir)

		// Verify symlinks were created
		targetFile1 := filepath.Join(mkdirTargetDir, "file1.txt")
		targetFile2 := filepath.Join(mkdirTargetDir, "file2.txt")
		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)
	})
}

func TestInstallForceMode(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	// Create source directory and files
	err := os.MkdirAll(sourceDir, 0755)
	require.NoError(t, err)

	sourceFile1 := filepath.Join(sourceDir, "file1.txt")
	sourceFile2 := filepath.Join(sourceDir, "file2.txt")

	err = os.WriteFile(sourceFile1, []byte("source1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sourceFile2, []byte("source2"), 0644)
	require.NoError(t, err)

	// Create target directory
	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create conflicting files in target directory
	targetFile1 := filepath.Join(targetDir, "file1.txt")
	targetFile2 := filepath.Join(targetDir, "file2.txt")

	err = os.WriteFile(targetFile1, []byte("existing1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(targetFile2, []byte("existing2"), 0644)
	require.NoError(t, err)

	modules := []config.ModuleConfig{
		{
			Dir:       sourceDir,
			TargetDir: targetDir,
		},
	}

	t.Run("force mode backs up existing files and creates symlinks", func(t *testing.T) {
		result, err := Install(modules, false, true) // force = true
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 2)
		assert.Len(t, result.SkippedLinks, 0)
		assert.Len(t, result.Errors, 0)

		// Verify backup files exist
		backupFile1 := filepath.Join(targetDir, "file1.txt.bak")
		backupFile2 := filepath.Join(targetDir, "file2.txt.bak")
		assert.FileExists(t, backupFile1)
		assert.FileExists(t, backupFile2)

		// Verify backup content
		backupContent1, err := os.ReadFile(backupFile1)
		require.NoError(t, err)
		assert.Equal(t, "existing1", string(backupContent1))

		backupContent2, err := os.ReadFile(backupFile2)
		require.NoError(t, err)
		assert.Equal(t, "existing2", string(backupContent2))

		// Verify symlinks were created
		assert.FileExists(t, targetFile1)
		assert.FileExists(t, targetFile2)

		linkTarget1, err := os.Readlink(targetFile1)
		require.NoError(t, err)
		absSource1, err := filepath.Abs(sourceFile1)
		require.NoError(t, err)
		assert.Equal(t, absSource1, linkTarget1)

		linkTarget2, err := os.Readlink(targetFile2)
		require.NoError(t, err)
		absSource2, err := filepath.Abs(sourceFile2)
		require.NoError(t, err)
		assert.Equal(t, absSource2, linkTarget2)
	})
}

func TestCreateSymlink(t *testing.T) {
	tempDir := t.TempDir()

	sourceFile := filepath.Join(tempDir, "source.txt")
	targetFile := filepath.Join(tempDir, "target.txt")

	err := os.WriteFile(sourceFile, []byte("test"), 0644)
	require.NoError(t, err)

	t.Run("create symlink successfully", func(t *testing.T) {
		err := createSymlink(sourceFile, targetFile, false)
		require.NoError(t, err)

		// Verify symlink exists and points to correct file using absolute path
		assert.FileExists(t, targetFile)
		linkTarget, err := os.Readlink(targetFile)
		require.NoError(t, err)

		// Should be absolute path
		absSource, err := filepath.Abs(sourceFile)
		require.NoError(t, err)
		assert.Equal(t, absSource, linkTarget)
	})

	t.Run("fail when target directory does not exist", func(t *testing.T) {
		nestedTarget := filepath.Join(tempDir, "nested", "dir", "target.txt")

		err := createSymlink(sourceFile, nestedTarget, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target directory does not exist")
	})

	t.Run("create directories when mkdir is true", func(t *testing.T) {
		nestedTarget := filepath.Join(tempDir, "mkdir", "nested", "dir", "target.txt")

		err := createSymlink(sourceFile, nestedTarget, true)
		require.NoError(t, err)

		// Verify symlink was created
		assert.FileExists(t, nestedTarget)
		linkTarget, err := os.Readlink(nestedTarget)
		require.NoError(t, err)

		absSource, err := filepath.Abs(sourceFile)
		require.NoError(t, err)
		assert.Equal(t, absSource, linkTarget)

		// Verify directories were created
		nestedDir := filepath.Join(tempDir, "mkdir", "nested", "dir")
		assert.DirExists(t, nestedDir)
	})
}

func TestBackupAndCreateSymlink(t *testing.T) {
	tempDir := t.TempDir()

	sourceFile := filepath.Join(tempDir, "source.txt")
	targetFile := filepath.Join(tempDir, "target.txt")

	err := os.WriteFile(sourceFile, []byte("source content"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(targetFile, []byte("existing content"), 0644)
	require.NoError(t, err)

	t.Run("backup existing file and create symlink", func(t *testing.T) {
		err := backupAndCreateSymlink(sourceFile, targetFile, false)
		require.NoError(t, err)

		// Verify backup file exists with correct content
		backupFile := targetFile + ".bak"
		assert.FileExists(t, backupFile)
		backupContent, err := os.ReadFile(backupFile)
		require.NoError(t, err)
		assert.Equal(t, "existing content", string(backupContent))

		// Verify symlink was created
		assert.FileExists(t, targetFile)
		linkTarget, err := os.Readlink(targetFile)
		require.NoError(t, err)

		absSource, err := filepath.Abs(sourceFile)
		require.NoError(t, err)
		assert.Equal(t, absSource, linkTarget)
	})

	t.Run("overwrite existing backup file", func(t *testing.T) {
		// Recreate the target file with different content (since previous test created a symlink)
		err := os.WriteFile(targetFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		// Create a backup file with different content
		backupFile := targetFile + ".bak"
		err = os.WriteFile(backupFile, []byte("old backup"), 0644)
		require.NoError(t, err)

		// Run backup again
		err = backupAndCreateSymlink(sourceFile, targetFile, false)
		require.NoError(t, err)

		// Verify backup was overwritten with the current target content
		backupContent, err := os.ReadFile(backupFile)
		require.NoError(t, err)
		assert.Equal(t, "existing content", string(backupContent))
	})
}

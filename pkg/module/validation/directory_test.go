package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirectoryValidator_ValidateTargetDirectories(t *testing.T) {
	tempDir := t.TempDir()

	validator := NewDirectoryValidator()

	t.Run("valid target directory", func(t *testing.T) {
		targetDir := filepath.Join(tempDir, "valid_target")
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		moduleConfig := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := validator.ValidateTargetDirectories([]config.ModuleConfig{moduleConfig}, false)
		assert.Empty(t, errors)
	})

	t.Run("target directory does not exist", func(t *testing.T) {
		targetDir := filepath.Join(tempDir, "nonexistent_target")

		moduleConfig := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := validator.ValidateTargetDirectories([]config.ModuleConfig{moduleConfig}, false)
		assert.NotEmpty(t, errors) // Non-existent directories should fail
		assert.Contains(t, errors[0].Error(), "target directory does not exist")
	})

	t.Run("target is a symlink", func(t *testing.T) {
		realDir := filepath.Join(tempDir, "real_dir")
		linkDir := filepath.Join(tempDir, "link_dir")

		err := os.MkdirAll(realDir, 0755)
		require.NoError(t, err)

		// Remove the existing file if it exists
		os.RemoveAll(linkDir)

		err = os.Symlink(realDir, linkDir)
		require.NoError(t, err)

		moduleConfig := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: linkDir,
		}

		errors := validator.ValidateTargetDirectories([]config.ModuleConfig{moduleConfig}, false)
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0].Error(), "is a symlink")
	})

	t.Run("parent is a symlink", func(t *testing.T) {
		realDir := filepath.Join(tempDir, "real_dir")
		linkDir := filepath.Join(tempDir, "link_dir")
		targetDir := filepath.Join(linkDir, "subdir")

		err := os.MkdirAll(realDir, 0755)
		require.NoError(t, err)

		// Create the subdir in the real directory so the target exists
		subDir := filepath.Join(realDir, "subdir")
		err = os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		// Remove the existing file if it exists
		os.RemoveAll(linkDir)

		err = os.Symlink(realDir, linkDir)
		require.NoError(t, err)

		moduleConfig := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := validator.ValidateTargetDirectories([]config.ModuleConfig{moduleConfig}, false)
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0].Error(), "is a symlink")
	})

	t.Run("target directory does not exist but mkdir is true", func(t *testing.T) {
		targetDir := filepath.Join(tempDir, "mkdir_target")

		moduleConfig := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := validator.ValidateTargetDirectories([]config.ModuleConfig{moduleConfig}, true)
		assert.Empty(t, errors) // Should not fail when mkdir is true
	})
}

func TestDirectoryValidator_ValidateDirectory(t *testing.T) {
	tempDir := t.TempDir()

	validator := NewDirectoryValidator()

	t.Run("valid directory structure", func(t *testing.T) {
		dir := filepath.Join(tempDir, "valid", "path")
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = validator.ValidateDirectory(dir, false)
		assert.NoError(t, err)
	})

	t.Run("non-existent directory", func(t *testing.T) {
		nonExistent := filepath.Join(tempDir, "nonexistent", "path")
		err := validator.ValidateDirectory(nonExistent, false)
		assert.Error(t, err) // Non-existent directories should fail
		assert.Contains(t, err.Error(), "target directory does not exist")
	})

	t.Run("directory is a symlink", func(t *testing.T) {
		realDir := filepath.Join(tempDir, "real")
		linkDir := filepath.Join(tempDir, "link")

		err := os.MkdirAll(realDir, 0755)
		require.NoError(t, err)

		// Remove the existing file if it exists
		os.RemoveAll(linkDir)

		err = os.Symlink(realDir, linkDir)
		require.NoError(t, err)

		err = validator.ValidateDirectory(linkDir, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is a symlink")
	})

	t.Run("regular file instead of directory", func(t *testing.T) {
		file := filepath.Join(tempDir, "file")
		err := os.WriteFile(file, []byte("content"), 0644)
		require.NoError(t, err)

		err = validator.ValidateDirectory(file, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a directory")
	})

	t.Run("non-existent directory with mkdir true", func(t *testing.T) {
		nonExistent := filepath.Join(tempDir, "mkdir", "path")

		err := validator.ValidateDirectory(nonExistent, true)
		assert.NoError(t, err) // Should not fail when mkdir is true
	})
}

func TestDirectoryError(t *testing.T) {
	err := DirectoryError{
		Module:  "test-module",
		Path:    "/test/path",
		Message: "test error message",
	}

	assert.Equal(t, "test error message", err.Error())
	assert.Equal(t, "test-module", err.Module)
	assert.Equal(t, "/test/path", err.Path)
}

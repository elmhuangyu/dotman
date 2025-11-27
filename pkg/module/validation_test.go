package module

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateFileMapping(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0644)
	require.NoError(t, err)

	t.Run("target does not exist", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")

		operation, err := validateFileMapping(sourceFile, targetFile)
		require.NoError(t, err)
		assert.Equal(t, OperationCreateLink, operation.Type)
		assert.Equal(t, sourceFile, operation.Source)
		assert.Equal(t, targetFile, operation.Target)
	})

	t.Run("target exists as correct symlink", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "correct_link.txt")

		// Create correct symlink
		err := os.Symlink(sourceFile, targetFile)
		require.NoError(t, err)

		operation, err := validateFileMapping(sourceFile, targetFile)
		require.NoError(t, err)
		assert.Equal(t, OperationSkip, operation.Type)
	})

	t.Run("target exists as wrong symlink", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "wrong_link.txt")
		wrongSource := filepath.Join(tempDir, "wrong_source.txt")

		// Create wrong source file
		err := os.WriteFile(wrongSource, []byte("wrong content"), 0644)
		require.NoError(t, err)

		// Create wrong symlink
		err = os.Symlink(wrongSource, targetFile)
		require.NoError(t, err)

		operation, err := validateFileMapping(sourceFile, targetFile)
		require.NoError(t, err)
		assert.Equal(t, OperationConflict, operation.Type)
		assert.Contains(t, operation.Description, "target exists as symlink pointing to wrong file")
		assert.Contains(t, operation.Description, wrongSource)
	})

	t.Run("target exists as regular file", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "regular_file.txt")

		// Create regular file at target location
		err := os.WriteFile(targetFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		operation, err := validateFileMapping(sourceFile, targetFile)
		require.NoError(t, err)
		assert.Equal(t, OperationConflict, operation.Type)
		assert.Equal(t, "target exists as regular file", operation.Description)
	})

	t.Run("source file does not exist", func(t *testing.T) {
		nonExistentSource := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		_, err := validateFileMapping(nonExistentSource, targetFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source file does not exist")
	})

	t.Run("source is directory", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "source_dir")
		err := os.Mkdir(sourceDir, 0755)
		require.NoError(t, err)

		targetFile := filepath.Join(tempDir, "target.txt")

		_, err = validateFileMapping(sourceDir, targetFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source is a directory")
	})
}

func TestValidateTargetDirectories(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid target directory", func(t *testing.T) {
		targetDir := filepath.Join(tempDir, "valid_target")
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		module := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := ValidateTargetDirectories([]config.ModuleConfig{module})
		assert.Empty(t, errors)
	})

	t.Run("target directory does not exist", func(t *testing.T) {
		targetDir := filepath.Join(tempDir, "nonexistent_target")

		module := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := ValidateTargetDirectories([]config.ModuleConfig{module})
		assert.NotEmpty(t, errors) // Non-existent directories should fail
		assert.Contains(t, errors[0], "target directory does not exist")
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

		module := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: linkDir,
		}

		errors := ValidateTargetDirectories([]config.ModuleConfig{module})
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "is a symlink")
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

		module := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := ValidateTargetDirectories([]config.ModuleConfig{module})
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "is a symlink")
	})
}

func TestValidateDirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid directory structure", func(t *testing.T) {
		dir := filepath.Join(tempDir, "valid", "path")
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = validateDirectoryStructure(dir)
		assert.NoError(t, err)
	})

	t.Run("non-existent directory", func(t *testing.T) {
		nonExistent := filepath.Join(tempDir, "nonexistent", "path")
		err := validateDirectoryStructure(nonExistent)
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

		err = validateDirectoryStructure(linkDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is a symlink")
	})

	t.Run("regular file instead of directory", func(t *testing.T) {
		file := filepath.Join(tempDir, "file")
		err := os.WriteFile(file, []byte("content"), 0644)
		require.NoError(t, err)

		err = validateDirectoryStructure(file)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a directory")
	})
}

func TestValidateInstallation(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("successful validation", func(t *testing.T) {
		// Create source files
		sourceDir := filepath.Join(tempDir, "source")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)

		sourceFile1 := filepath.Join(sourceDir, "file1.txt")
		sourceFile2 := filepath.Join(sourceDir, "file2.txt")
		err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
		require.NoError(t, err)

		// Create target directory (must exist for successful validation)
		targetDir := filepath.Join(tempDir, "target")
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create module config
		module := config.ModuleConfig{
			Dir:       sourceDir,
			TargetDir: targetDir,
		}

		validation, err := ValidateInstallation([]config.ModuleConfig{module})
		require.NoError(t, err)
		assert.NotNil(t, validation)
		assert.True(t, validation.IsValid)
		assert.Empty(t, validation.Errors)

		// Should have 2 create operations
		createOps := 0
		for _, op := range validation.Operations {
			if op.Type == OperationCreateLink {
				createOps++
			}
		}
		assert.Equal(t, 2, createOps)
	})

	t.Run("validation with conflicts", func(t *testing.T) {
		// Create two modules that would map to the same target
		sourceDir1 := filepath.Join(tempDir, "source1")
		sourceDir2 := filepath.Join(tempDir, "source2")
		err := os.MkdirAll(sourceDir1, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(sourceDir2, 0755)
		require.NoError(t, err)

		sourceFile1 := filepath.Join(sourceDir1, "config.txt")
		sourceFile2 := filepath.Join(sourceDir2, "config.txt")
		err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
		require.NoError(t, err)

		// Both modules target the same directory
		targetDir := filepath.Join(tempDir, "target")
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		module1 := config.ModuleConfig{
			Dir:       sourceDir1,
			TargetDir: targetDir,
		}
		module2 := config.ModuleConfig{
			Dir:       sourceDir2,
			TargetDir: targetDir,
		}

		validation, err := ValidateInstallation([]config.ModuleConfig{module1, module2})
		require.NoError(t, err)
		assert.NotNil(t, validation)
		assert.False(t, validation.IsValid)
		assert.NotEmpty(t, validation.Errors)

		// Should have conflict errors
		hasConflict := false
		for _, err := range validation.Errors {
			if strings.Contains(err, "target conflict") {
				hasConflict = true
				break
			}
		}
		assert.True(t, hasConflict)
	})
}


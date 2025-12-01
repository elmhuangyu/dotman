package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/module"
	"github.com/elmhuangyu/dotman/pkg/module/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_ValidateFileMapping(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create template renderer
	templateRenderer := template.NewRenderer()
	var tr template.TemplateRenderer = templateRenderer
	validator := NewValidator(tr)

	t.Run("target does not exist", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")

		operation, err := validator.validateFileMapping(sourceFile, targetFile, false, map[string]string{})
		require.NoError(t, err)
		assert.Equal(t, module.OperationCreateLink, operation.Type)
		assert.Equal(t, sourceFile, operation.Source)
		assert.Equal(t, targetFile, operation.Target)
	})

	t.Run("target exists as correct symlink", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "correct_link.txt")

		// Create correct symlink
		err := os.Symlink(sourceFile, targetFile)
		require.NoError(t, err)

		operation, err := validator.validateFileMapping(sourceFile, targetFile, false, map[string]string{})
		require.NoError(t, err)
		assert.Equal(t, module.OperationSkip, operation.Type)
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

		operation, err := validator.validateFileMapping(sourceFile, targetFile, false, map[string]string{})
		require.NoError(t, err)
		assert.Equal(t, module.OperationForceLink, operation.Type)
		assert.Contains(t, operation.Description, "target exists as symlink pointing to wrong file")
		assert.Contains(t, operation.Description, wrongSource)
	})

	t.Run("target exists as regular file", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "regular_file.txt")

		// Create regular file at target location
		err := os.WriteFile(targetFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		operation, err := validator.validateFileMapping(sourceFile, targetFile, false, map[string]string{})
		require.NoError(t, err)
		assert.Equal(t, module.OperationForceLink, operation.Type)
		assert.Equal(t, "target exists as regular file", operation.Description)
	})

	t.Run("target exists as regular file (template)", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "template_target.txt")

		// Create regular file at target location
		err := os.WriteFile(targetFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		operation, err := validator.validateFileMapping(sourceFile, targetFile, true, map[string]string{})
		require.NoError(t, err)
		assert.Equal(t, module.OperationForceTemplate, operation.Type)
		assert.Equal(t, "target exists as file (template would overwrite)", operation.Description)
	})

	t.Run("source file does not exist", func(t *testing.T) {
		nonExistentSource := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		_, err := validator.validateFileMapping(nonExistentSource, targetFile, false, map[string]string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source file does not exist")
	})

	t.Run("source is directory", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "source_dir")
		err := os.Mkdir(sourceDir, 0755)
		require.NoError(t, err)

		targetFile := filepath.Join(tempDir, "target.txt")

		_, err = validator.validateFileMapping(sourceDir, targetFile, false, map[string]string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source is a directory")
	})
}

func TestValidator_ValidateTargetDirectories(t *testing.T) {
	tempDir := t.TempDir()

	templateRenderer := template.NewRenderer()
	validator := NewValidator(templateRenderer)

	t.Run("valid target directory", func(t *testing.T) {
		targetDir := filepath.Join(tempDir, "valid_target")
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		module := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := validator.ValidateTargetDirectories([]config.ModuleConfig{module}, false)
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

		moduleConfig := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: linkDir,
		}

		errors := validator.ValidateTargetDirectories([]config.ModuleConfig{moduleConfig}, false)
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

		moduleConfig := config.ModuleConfig{
			Dir:       filepath.Join(tempDir, "module"),
			TargetDir: targetDir,
		}

		errors := validator.ValidateTargetDirectories([]config.ModuleConfig{moduleConfig}, false)
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "is a symlink")
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

func TestValidator_ValidateDirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()

	templateRenderer := template.NewRenderer()
	validator := NewValidator(templateRenderer)

	t.Run("valid directory structure", func(t *testing.T) {
		dir := filepath.Join(tempDir, "valid", "path")
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = validator.validateDirectoryStructure(dir, false)
		assert.NoError(t, err)
	})

	t.Run("non-existent directory", func(t *testing.T) {
		nonExistent := filepath.Join(tempDir, "nonexistent", "path")
		err := validator.validateDirectoryStructure(nonExistent, false)
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

		err = validator.validateDirectoryStructure(linkDir, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is a symlink")
	})

	t.Run("regular file instead of directory", func(t *testing.T) {
		file := filepath.Join(tempDir, "file")
		err := os.WriteFile(file, []byte("content"), 0644)
		require.NoError(t, err)

		err = validator.validateDirectoryStructure(file, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a directory")
	})

	t.Run("non-existent directory with mkdir true", func(t *testing.T) {
		nonExistent := filepath.Join(tempDir, "mkdir", "path")

		err := validator.validateDirectoryStructure(nonExistent, true)
		assert.NoError(t, err) // Should not fail when mkdir is true
	})
}

func TestValidator_ValidateInstallation(t *testing.T) {
	tempDir := t.TempDir()

	templateRenderer := template.NewRenderer()
	validator := NewValidator(templateRenderer)

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

		// Create module directory and files
		moduleDir := filepath.Join(tempDir, "module")
		err = os.MkdirAll(moduleDir, 0755)
		require.NoError(t, err)

		// Copy source files to module directory
		moduleFile1 := filepath.Join(moduleDir, "file1.txt")
		moduleFile2 := filepath.Join(moduleDir, "file2.txt")
		err = os.WriteFile(moduleFile1, []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(moduleFile2, []byte("content2"), 0644)
		require.NoError(t, err)

		// Create module config
		moduleConfig := config.ModuleConfig{
			Dir:       moduleDir,
			TargetDir: targetDir,
		}

		validation, err := validator.ValidateInstallation([]config.ModuleConfig{moduleConfig}, map[string]string{})
		require.NoError(t, err)
		assert.NotNil(t, validation)
		assert.True(t, validation.IsValid)
		assert.Empty(t, validation.Errors)

		// Should have 2 create operations
		createOps := 0
		for _, op := range validation.Operations {
			if op.Type == module.OperationCreateLink {
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

		validation, err := validator.ValidateInstallation([]config.ModuleConfig{module1, module2}, map[string]string{})
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

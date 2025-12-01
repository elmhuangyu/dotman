package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/module"
	"github.com/elmhuangyu/dotman/pkg/module/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMappingValidator_ValidateFileMapping(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create template renderer
	templateRenderer := template.NewRenderer()
	validator := NewMappingValidator(templateRenderer)

	t.Run("target does not exist", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")

		operation, err := validator.ValidateFileMapping(sourceFile, targetFile, false, map[string]string{})
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

		operation, err := validator.ValidateFileMapping(sourceFile, targetFile, false, map[string]string{})
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

		operation, err := validator.ValidateFileMapping(sourceFile, targetFile, false, map[string]string{})
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

		operation, err := validator.ValidateFileMapping(sourceFile, targetFile, false, map[string]string{})
		require.NoError(t, err)
		assert.Equal(t, module.OperationForceLink, operation.Type)
		assert.Equal(t, "target exists as regular file", operation.Description)
	})

	t.Run("target exists as regular file (template)", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "template_target.txt")

		// Create regular file at target location
		err := os.WriteFile(targetFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		operation, err := validator.ValidateFileMapping(sourceFile, targetFile, true, map[string]string{})
		require.NoError(t, err)
		assert.Equal(t, module.OperationForceTemplate, operation.Type)
		assert.Equal(t, "target exists as file (template would overwrite)", operation.Description)
	})

	t.Run("source file does not exist", func(t *testing.T) {
		nonExistentSource := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		_, err := validator.ValidateFileMapping(nonExistentSource, targetFile, false, map[string]string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source file does not exist")
	})

	t.Run("source is directory", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "source_dir")
		err := os.Mkdir(sourceDir, 0755)
		require.NoError(t, err)

		targetFile := filepath.Join(tempDir, "target.txt")

		_, err = validator.ValidateFileMapping(sourceDir, targetFile, false, map[string]string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source is a directory")
	})
}

func TestMappingValidator_DetectConflicts(t *testing.T) {
	templateRenderer := template.NewRenderer()
	validator := NewMappingValidator(templateRenderer)

	t.Run("no conflicts", func(t *testing.T) {
		mapping := module.NewFileMapping()
		mapping.AddMapping("/source1", "/target1")
		mapping.AddMapping("/source2", "/target2")

		conflicts := validator.DetectConflicts(mapping)
		assert.Empty(t, conflicts)
	})

	t.Run("has conflicts", func(t *testing.T) {
		mapping := module.NewFileMapping()
		mapping.AddMapping("/source1", "/target")
		mapping.AddMapping("/source2", "/target")

		conflicts := validator.DetectConflicts(mapping)
		assert.Len(t, conflicts, 1)
		assert.Equal(t, "/target", conflicts[0].Target)
		assert.Contains(t, conflicts[0].Sources, "/source1")
		assert.Contains(t, conflicts[0].Sources, "/source2")
		assert.Contains(t, conflicts[0].Message, "target conflict")
	})
}

func TestConflictError(t *testing.T) {
	conflict := ConflictError{
		Target:  "/target",
		Sources: []string{"/source1", "/source2"},
		Message: "test conflict",
	}

	assert.Equal(t, "test conflict", conflict.Error())
	assert.Equal(t, "/target", conflict.Target)
	assert.Equal(t, []string{"/source1", "/source2"}, conflict.Sources)
}

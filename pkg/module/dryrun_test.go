package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDryRun(t *testing.T) {
	t.Run("successful dry run with no existing files", func(t *testing.T) {
		// Create source files
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)

		sourceFile1 := filepath.Join(sourceDir, "file1.txt")
		sourceFile2 := filepath.Join(sourceDir, "file2.txt")
		err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
		require.NoError(t, err)

		// Create target directory (this should exist for successful validation)
		targetDir := filepath.Join(tempDir, "target")
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create module config
		module := config.ModuleConfig{
			Dir:       sourceDir,
			TargetDir: targetDir,
		}

		result, err := Validate([]config.ModuleConfig{module}, map[string]string{}, false, false)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
		assert.Len(t, result.CreateOperations, 2)
		assert.Empty(t, result.SkipOperations)
		assert.Empty(t, result.ForceLinkOperations)
		assert.Empty(t, result.ForceTemplateOps)

		// Check summary
		assert.Contains(t, result.Summary, "2 total file operations")
		assert.Contains(t, result.Summary, "2 files would be linked")
	})

	t.Run("dry run with existing correct symlinks", func(t *testing.T) {
		// Create source files
		tempDir := t.TempDir()

		sourceDir := filepath.Join(tempDir, "source")
		targetDir := filepath.Join(tempDir, "target")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		sourceFile := filepath.Join(sourceDir, "config.txt")
		targetFile := filepath.Join(targetDir, "config.txt")
		err = os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		// Create Dotfile config for module
		dotfileContent := `target_dir: "` + targetDir + `"`
		err = os.WriteFile(filepath.Join(sourceDir, "Dotfile"), []byte(dotfileContent), 0644)
		require.NoError(t, err)

		// Create correct symlink
		err = os.Symlink(sourceFile, targetFile)
		require.NoError(t, err)

		// Load module config
		moduleConfig, err := config.LoadConfig(sourceDir)
		require.NoError(t, err)
		require.NotNil(t, moduleConfig)

		result, err := Validate([]config.ModuleConfig{*moduleConfig}, map[string]string{}, false, false)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
		assert.Len(t, result.CreateOperations, 0)
		assert.Len(t, result.SkipOperations, 1)
		assert.Empty(t, result.ForceLinkOperations)
		assert.Empty(t, result.ForceTemplateOps)

		// Check summary
		assert.Contains(t, result.Summary, "1 files skipped")
	})

	t.Run("dry run with conflicts", func(t *testing.T) {
		// Create source file
		tempDir := t.TempDir()

		sourceDir := filepath.Join(tempDir, "source")
		targetDir := filepath.Join(tempDir, "target")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		sourceFile := filepath.Join(sourceDir, "config.txt")
		targetFile := filepath.Join(targetDir, "config.txt")
		err = os.WriteFile(sourceFile, []byte("new content"), 0644)
		require.NoError(t, err)

		// Create Dotfile config for module
		dotfileContent := `target_dir: "` + targetDir + `"`
		err = os.WriteFile(filepath.Join(sourceDir, "Dotfile"), []byte(dotfileContent), 0644)
		require.NoError(t, err)

		// Create existing file at target location
		err = os.WriteFile(targetFile, []byte("old content"), 0644)
		require.NoError(t, err)

		// Load module config
		moduleConfig, err := config.LoadConfig(sourceDir)
		require.NoError(t, err)
		require.NotNil(t, moduleConfig)

		result, err := Validate([]config.ModuleConfig{*moduleConfig}, map[string]string{}, false, false)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.False(t, result.IsValid) // Should be invalid due to conflict
		assert.Empty(t, result.Errors)  // No validation errors, but conflict operations exist
		assert.Empty(t, result.CreateOperations)
		assert.Empty(t, result.SkipOperations)
		assert.Len(t, result.ForceLinkOperations, 1)
		assert.Empty(t, result.ForceTemplateOps)

		// Check summary
		assert.Contains(t, result.Summary, "1 conflicts found")
	})

	t.Run("dry run with wrong symlinks", func(t *testing.T) {
		// Create source files
		tempDir := t.TempDir()

		sourceDir := filepath.Join(tempDir, "source")
		targetDir := filepath.Join(tempDir, "target")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		sourceFile := filepath.Join(sourceDir, "config.txt")
		wrongSource := filepath.Join(sourceDir, "old_config.txt")
		targetFile := filepath.Join(targetDir, "config.txt")
		err = os.WriteFile(sourceFile, []byte("new content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(wrongSource, []byte("old content"), 0644)
		require.NoError(t, err)

		// Create Dotfile config for module
		dotfileContent := `target_dir: "` + targetDir + `"`
		err = os.WriteFile(filepath.Join(sourceDir, "Dotfile"), []byte(dotfileContent), 0644)
		require.NoError(t, err)

		// Create wrong symlink
		err = os.Symlink(wrongSource, targetFile)
		require.NoError(t, err)

		// Load module config
		moduleConfig, err := config.LoadConfig(sourceDir)
		require.NoError(t, err)
		require.NotNil(t, moduleConfig)

		result, err := Validate([]config.ModuleConfig{*moduleConfig}, map[string]string{}, false, false)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.False(t, result.IsValid) // Wrong symlinks are now conflicts
		assert.Empty(t, result.Errors)
		assert.Len(t, result.CreateOperations, 1) // old_config.txt
		assert.Empty(t, result.SkipOperations)
		assert.Len(t, result.ForceLinkOperations, 1) // config.txt (wrong symlink)
		assert.Empty(t, result.ForceTemplateOps)

		// Check summary
		assert.Contains(t, result.Summary, "2 total file operations")
		assert.Contains(t, result.Summary, "1 files would be linked")
		assert.Contains(t, result.Summary, "1 conflicts found")
	})

	t.Run("dry run with missing target directory and mkdir enabled", func(t *testing.T) {
		// Create source files
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)

		sourceFile1 := filepath.Join(sourceDir, "file1.txt")
		sourceFile2 := filepath.Join(sourceDir, "file2.txt")
		err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
		require.NoError(t, err)

		// Use target directory that doesn't exist
		targetDir := filepath.Join(tempDir, "missing", "target", "dir")

		// Create module config
		module := config.ModuleConfig{
			Dir:       sourceDir,
			TargetDir: targetDir,
		}

		result, err := Validate([]config.ModuleConfig{module}, map[string]string{}, true, false)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
		assert.Len(t, result.CreateOperations, 2)
		assert.Empty(t, result.SkipOperations)
		assert.Empty(t, result.ForceLinkOperations)
		assert.Empty(t, result.ForceTemplateOps)

		// Check summary
		assert.Contains(t, result.Summary, "2 total file operations")
		assert.Contains(t, result.Summary, "2 files would be linked")
	})
}

func TestGenerateDryRunSummary(t *testing.T) {
	result := &ValidateResult{
		CreateOperations:    []FileOperation{{Type: OperationCreateLink}},
		SkipOperations:      []FileOperation{{Type: OperationSkip}, {Type: OperationSkip}, {Type: OperationSkip}},
		ForceLinkOperations: []FileOperation{{Type: OperationForceLink}},
		Errors:              []string{"error1"},
	}

	summary := generateValidationSummary(result, false)

	assert.Contains(t, summary, "5 total file operations")
	assert.Contains(t, summary, "1 files would be linked")
	assert.Contains(t, summary, "3 files skipped")
	assert.Contains(t, summary, "1 conflicts found")
	assert.Contains(t, summary, "1 errors")
}

func TestSortFileOperations(t *testing.T) {
	operations := []FileOperation{
		{Type: OperationCreateLink, Target: "/z.txt"},
		{Type: OperationCreateLink, Target: "/a.txt"},
		{Type: OperationCreateLink, Target: "/m.txt"},
	}

	sortFileOperations(operations)

	assert.Equal(t, "/a.txt", operations[0].Target)
	assert.Equal(t, "/m.txt", operations[1].Target)
	assert.Equal(t, "/z.txt", operations[2].Target)
}

func TestDryRunWithComplexSetup(t *testing.T) {
	tempDir := t.TempDir()

	// Create complex setup with multiple modules and various file states
	sourceDir1 := filepath.Join(tempDir, "source1")
	sourceDir2 := filepath.Join(tempDir, "source2")
	targetDir1 := filepath.Join(tempDir, "target1")
	targetDir2 := filepath.Join(tempDir, "target2")

	// Create directories
	err := os.MkdirAll(sourceDir1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(sourceDir2, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(targetDir1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(targetDir2, 0755)
	require.NoError(t, err)

	// Module 1: New files and correct symlinks
	sourceFile1 := filepath.Join(sourceDir1, "new.txt")
	sourceFile2 := filepath.Join(sourceDir1, "correct.txt")
	targetFile2 := filepath.Join(targetDir1, "correct.txt")
	err = os.WriteFile(sourceFile1, []byte("new content"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sourceFile2, []byte("correct content"), 0644)
	require.NoError(t, err)
	err = os.Symlink(sourceFile2, targetFile2)
	require.NoError(t, err)

	// Module 2: Wrong symlinks and conflicts
	sourceFile3 := filepath.Join(sourceDir2, "update.txt")
	sourceFile4 := filepath.Join(sourceDir2, "conflict.txt")
	wrongSource := filepath.Join(sourceDir2, "old.txt")
	targetFile3 := filepath.Join(targetDir2, "update.txt")
	targetFile4 := filepath.Join(targetDir2, "conflict.txt")
	err = os.WriteFile(sourceFile3, []byte("update content"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sourceFile4, []byte("conflict content"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(wrongSource, []byte("old content"), 0644)
	require.NoError(t, err)
	err = os.Symlink(wrongSource, targetFile3) // Wrong symlink
	require.NoError(t, err)
	err = os.WriteFile(targetFile4, []byte("existing content"), 0644) // Conflict file
	require.NoError(t, err)

	// Create module configs
	module1 := config.ModuleConfig{
		Dir:       sourceDir1,
		TargetDir: targetDir1,
	}
	module2 := config.ModuleConfig{
		Dir:       sourceDir2,
		TargetDir: targetDir2,
	}

	result, err := Validate([]config.ModuleConfig{module1, module2}, map[string]string{}, false, false)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.IsValid)              // Due to conflicts
	assert.Len(t, result.CreateOperations, 2)    // new.txt + old.txt
	assert.Len(t, result.SkipOperations, 1)      // correct.txt (already correct)
	assert.Len(t, result.ForceLinkOperations, 2) // update.txt (wrong symlink) + conflict.txt (existing file)
	assert.Len(t, result.ForceTemplateOps, 0)

	// Verify summary
	assert.Contains(t, result.Summary, "5 total file operations")
	assert.Contains(t, result.Summary, "2 files would be linked")
	assert.Contains(t, result.Summary, "1 files skipped")
	assert.Contains(t, result.Summary, "2 conflicts found")
}

func TestDryRunLogOutput(t *testing.T) {
	// This test doesn't actually test logging output, but ensures the function doesn't panic
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")
	err := os.MkdirAll(sourceDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	sourceFile := filepath.Join(sourceDir, "test.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0644)
	require.NoError(t, err)

	module := config.ModuleConfig{
		Dir:       sourceDir,
		TargetDir: targetDir,
	}

	result, err := Validate([]config.ModuleConfig{module}, map[string]string{}, false, false)
	require.NoError(t, err)

	// This should not panic
	assert.NotPanics(t, func() {
		LogValidateResult(result)
	})
}

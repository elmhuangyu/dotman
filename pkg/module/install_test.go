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

		result, err := Install(modules, map[string]string{}, false, false, "")
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

		result, err := Install(modules, map[string]string{}, false, false, "")
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

		result, err := Install(modules, map[string]string{}, false, false, "")
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

		result, err := Install(nestedModules, map[string]string{}, false, false, "")
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

		result, err := Install(mkdirModules, map[string]string{}, true, false, "")
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
		result, err := Install(modules, map[string]string{}, false, true, "") // force = true
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

func TestInstallStateFileHandling(t *testing.T) {
	tempDir := t.TempDir()

	// Create test module structure
	moduleDir := filepath.Join(tempDir, "module")
	targetDir := filepath.Join(tempDir, "target")
	dotfilesDir := filepath.Join(tempDir, "dotfiles")

	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(dotfilesDir, 0755)
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

	t.Run("skipped files are recorded in state file", func(t *testing.T) {
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

		// Run installation with state file
		result, err := Install(modules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.SkippedLinks, 2)

		// Check state file contains skipped files
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile)

		// Verify both skipped files are in state file
		assert.Len(t, stateFile.Files, 2)

		absSource1, err := filepath.Abs(sourceFile1)
		require.NoError(t, err)
		absTarget1, err := filepath.Abs(targetFile1)
		require.NoError(t, err)

		absSource2, err := filepath.Abs(sourceFile2)
		require.NoError(t, err)
		absTarget2, err := filepath.Abs(targetFile2)
		require.NoError(t, err)

		// Check file mappings
		found1 := false
		found2 := false
		for _, file := range stateFile.Files {
			if file.Source == absSource1 && file.Target == absTarget1 && file.Type == state.TypeLink {
				found1 = true
			}
			if file.Source == absSource2 && file.Target == absTarget2 && file.Type == state.TypeLink {
				found2 = true
			}
		}
		assert.True(t, found1, "file1.txt mapping not found in state file")
		assert.True(t, found2, "file2.txt mapping not found in state file")
	})

	t.Run("state file avoids duplicates for skipped files", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		// Pre-create correct symlinks for both files
		targetFile1 := filepath.Join(targetDir, "file1.txt")
		targetFile2 := filepath.Join(targetDir, "file2.txt")
		err := os.Symlink(sourceFile1, targetFile1)
		require.NoError(t, err)
		err = os.Symlink(sourceFile2, targetFile2)
		require.NoError(t, err)

		// Run installation twice to test duplicate handling
		result1, err := Install(modules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result1.IsSuccess)
		assert.Len(t, result1.SkippedLinks, 2)

		result2, err := Install(modules, map[string]string{}, false, false, dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result2.IsSuccess)
		assert.Len(t, result2.SkippedLinks, 2)

		// Check state file contains only two entries (no duplicates)
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile)

		assert.Len(t, stateFile.Files, 2, "state file should not contain duplicate entries")
	})
}

func TestInstallTemplateFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test module structure
	moduleDir := filepath.Join(tempDir, "module")
	targetDir := filepath.Join(tempDir, "target")

	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create test files in module
	templateFile := filepath.Join(moduleDir, "config.dot-tmpl")
	regularFile := filepath.Join(moduleDir, "file1.txt")

	// Create template content
	templateContent := "User: {{.USER}}, Home: {{.HOME}}"
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	require.NoError(t, err)

	err = os.WriteFile(regularFile, []byte("regular content"), 0644)
	require.NoError(t, err)

	// Create module config
	modules := []config.ModuleConfig{
		{
			Dir:       moduleDir,
			TargetDir: targetDir,
			Ignores:   []string{},
		},
	}

	t.Run("successful template installation", func(t *testing.T) {
		// Clean target directory
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		vars := map[string]string{
			"USER": "testuser",
			"HOME": "/home/testuser",
		}

		result, err := Install(modules, vars, false, false, "")
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 1)     // regular file
		assert.Len(t, result.CreatedTemplates, 1) // template file
		assert.Len(t, result.SkippedLinks, 0)
		assert.Len(t, result.Errors, 0)
		assert.Contains(t, result.Summary, "1 symlinks created")
		assert.Contains(t, result.Summary, "1 template files generated")

		// Verify symlink was created for regular file
		targetRegularFile := filepath.Join(targetDir, "file1.txt")
		assert.FileExists(t, targetRegularFile)
		linkTarget, err := os.Readlink(targetRegularFile)
		require.NoError(t, err)
		absRegularSource, err := filepath.Abs(regularFile)
		require.NoError(t, err)
		assert.Equal(t, absRegularSource, linkTarget)

		// Verify template file was generated
		targetTemplateFile := filepath.Join(targetDir, "config")
		assert.FileExists(t, targetTemplateFile)

		// Check that it's not a symlink
		info, err := os.Lstat(targetTemplateFile)
		require.NoError(t, err)
		assert.False(t, info.Mode()&os.ModeSymlink != 0, "template file should not be a symlink")

		// Check content
		content, err := os.ReadFile(targetTemplateFile)
		require.NoError(t, err)
		assert.Equal(t, "User: testuser, Home: /home/testuser", string(content))
	})
}

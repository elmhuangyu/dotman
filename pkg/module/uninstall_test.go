package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUninstall(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	targetDir := filepath.Join(tempDir, "target")

	// Create directories
	err := os.MkdirAll(dotfilesDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create source files
	sourceFile1 := filepath.Join(dotfilesDir, "file1.txt")
	sourceFile2 := filepath.Join(dotfilesDir, "file2.txt")

	err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
	require.NoError(t, err)

	// Create target symlinks
	targetFile1 := filepath.Join(targetDir, "file1.txt")
	targetFile2 := filepath.Join(targetDir, "file2.txt")

	absSource1, err := filepath.Abs(sourceFile1)
	require.NoError(t, err)

	absSource2, err := filepath.Abs(sourceFile2)
	require.NoError(t, err)

	err = os.Symlink(absSource1, targetFile1)
	require.NoError(t, err)

	err = os.Symlink(absSource2, targetFile2)
	require.NoError(t, err)

	t.Run("successful uninstall with valid state file", func(t *testing.T) {
		// Create state file with the symlinks
		stateFile := state.NewStateFile()
		stateFile.AddFileMapping(sourceFile1, targetFile1, state.TypeLink)
		stateFile.AddFileMapping(sourceFile2, targetFile2, state.TypeLink)

		statePath := filepath.Join(dotfilesDir, "state.yaml")
		err := state.SaveStateFile(statePath, stateFile)
		require.NoError(t, err)

		// Run uninstall
		result, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.RemovedLinks, 2)
		assert.Len(t, result.SkippedLinks, 0)
		assert.Len(t, result.FailedRemovals, 0)
		assert.Len(t, result.Errors, 0)
		assert.Contains(t, result.Summary, "2 symlinks removed")

		// Verify symlinks are removed
		assert.NoFileExists(t, targetFile1)
		assert.NoFileExists(t, targetFile2)

		// Verify state file is updated (should be empty)
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 0)
	})

	t.Run("uninstall with missing state file", func(t *testing.T) {
		// Remove state file
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		os.Remove(statePath)

		// Run uninstall
		result, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.RemovedLinks, 0)
		assert.Len(t, result.SkippedLinks, 0)
		assert.Len(t, result.FailedRemovals, 0)
		assert.Contains(t, result.Summary, "No tracked installations found")
	})

	t.Run("uninstall with corrupted state file", func(t *testing.T) {
		// Create corrupted state file
		statePath := filepath.Join(dotfilesDir, "state.yaml")
		err := os.WriteFile(statePath, []byte("invalid: yaml: content"), 0644)
		require.NoError(t, err)

		// Run uninstall
		result, err := Uninstall(dotfilesDir)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to load state file")
	})
}

func TestValidateSymlink(t *testing.T) {
	tempDir := t.TempDir()

	sourceFile := filepath.Join(tempDir, "source.txt")
	targetFile := filepath.Join(tempDir, "target.txt")

	err := os.WriteFile(sourceFile, []byte("content"), 0644)
	require.NoError(t, err)

	absSource, err := filepath.Abs(sourceFile)
	require.NoError(t, err)

	t.Run("valid symlink", func(t *testing.T) {
		err := os.Symlink(absSource, targetFile)
		require.NoError(t, err)

		fileMapping := state.FileMapping{
			Source: sourceFile,
			Target: targetFile,
			Type:   state.TypeLink,
		}

		result := validateSymlink(fileMapping, logger.GetLogger())
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Reason)

		// Cleanup
		os.Remove(targetFile)
	})

	t.Run("target does not exist", func(t *testing.T) {
		fileMapping := state.FileMapping{
			Source: sourceFile,
			Target: filepath.Join(tempDir, "nonexistent.txt"),
			Type:   state.TypeLink,
		}

		result := validateSymlink(fileMapping, logger.GetLogger())
		assert.False(t, result.IsValid)
		assert.Contains(t, result.Reason, "target file does not exist")
	})

	t.Run("target is not a symlink", func(t *testing.T) {
		// Create a regular file
		err := os.WriteFile(targetFile, []byte("regular file"), 0644)
		require.NoError(t, err)

		fileMapping := state.FileMapping{
			Source: sourceFile,
			Target: targetFile,
			Type:   state.TypeLink,
		}

		result := validateSymlink(fileMapping, logger.GetLogger())
		assert.False(t, result.IsValid)
		assert.Contains(t, result.Reason, "target exists but is not a symlink")

		// Cleanup
		os.Remove(targetFile)
	})

	t.Run("symlink points to wrong target", func(t *testing.T) {
		wrongSource := filepath.Join(tempDir, "wrong.txt")
		err := os.WriteFile(wrongSource, []byte("wrong content"), 0644)
		require.NoError(t, err)

		err = os.Symlink(wrongSource, targetFile)
		require.NoError(t, err)

		fileMapping := state.FileMapping{
			Source: sourceFile,
			Target: targetFile,
			Type:   state.TypeLink,
		}

		result := validateSymlink(fileMapping, logger.GetLogger())
		assert.False(t, result.IsValid)
		assert.Contains(t, result.Reason, "symlink points to")

		// Cleanup
		os.Remove(targetFile)
		os.Remove(wrongSource)
	})

	t.Run("relative symlink", func(t *testing.T) {
		// Create a relative symlink
		relSource := "source.txt"
		err := os.Symlink(relSource, targetFile)
		require.NoError(t, err)

		fileMapping := state.FileMapping{
			Source: sourceFile,
			Target: targetFile,
			Type:   state.TypeLink,
		}

		result := validateSymlink(fileMapping, logger.GetLogger())
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Reason)

		// Cleanup
		os.Remove(targetFile)
	})
}

func TestRemoveSymlink(t *testing.T) {
	tempDir := t.TempDir()

	sourceFile := filepath.Join(tempDir, "source.txt")
	targetFile := filepath.Join(tempDir, "target.txt")

	err := os.WriteFile(sourceFile, []byte("content"), 0644)
	require.NoError(t, err)

	err = os.Symlink(sourceFile, targetFile)
	require.NoError(t, err)

	t.Run("remove existing symlink", func(t *testing.T) {
		err := removeSymlink(targetFile, logger.GetLogger())
		assert.NoError(t, err)

		// Verify symlink is removed
		assert.NoFileExists(t, targetFile)
	})

	t.Run("remove non-existent symlink", func(t *testing.T) {
		err := removeSymlink(filepath.Join(tempDir, "nonexistent.txt"), logger.GetLogger())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove symlink")
	})
}

func TestUpdateStateFile(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "state.yaml")

	// Create initial state file with multiple files
	stateFile := state.NewStateFile()
	stateFile.AddFileMapping("/source1", "/target1", state.TypeLink)
	stateFile.AddFileMapping("/source2", "/target2", state.TypeLink)
	stateFile.AddFileMapping("/source3", "/target3", state.TypeLink)

	err := state.SaveStateFile(statePath, stateFile)
	require.NoError(t, err)

	t.Run("remove some files from state file", func(t *testing.T) {
		// Create removed links for target1 and target3
		removedLinks := []FileOperation{
			{
				Type:   OperationCreateLink,
				Source: "/source1",
				Target: "/target1",
			},
			{
				Type:   OperationCreateLink,
				Source: "/source3",
				Target: "/target3",
			},
		}

		err := updateStateFile(statePath, stateFile, removedLinks, logger.GetLogger())
		assert.NoError(t, err)

		// Verify state file is updated
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)

		assert.Len(t, updatedStateFile.Files, 1)
		assert.Equal(t, "/source2", updatedStateFile.Files[0].Source)
		assert.Equal(t, "/target2", updatedStateFile.Files[0].Target)
	})

	t.Run("remove all files from state file", func(t *testing.T) {
		// Create removed links for remaining file
		removedLinks := []FileOperation{
			{
				Type:   OperationCreateLink,
				Source: "/source2",
				Target: "/target2",
			},
		}

		err := updateStateFile(statePath, stateFile, removedLinks, logger.GetLogger())
		assert.NoError(t, err)

		// Verify state file is empty
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)

		assert.Len(t, updatedStateFile.Files, 0)
	})
}

func TestUninstallMixedScenarios(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	targetDir := filepath.Join(tempDir, "target")

	// Create directories
	err := os.MkdirAll(dotfilesDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// Create source files
	sourceFile1 := filepath.Join(dotfilesDir, "file1.txt")
	sourceFile2 := filepath.Join(dotfilesDir, "file2.txt")
	sourceFile3 := filepath.Join(dotfilesDir, "file3.txt")

	err = os.WriteFile(sourceFile1, []byte("content1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(sourceFile2, []byte("content2"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(sourceFile3, []byte("content3"), 0644)
	require.NoError(t, err)

	// Create target symlinks for some files
	targetFile1 := filepath.Join(targetDir, "file1.txt")
	targetFile2 := filepath.Join(targetDir, "file2.txt")
	targetFile3 := filepath.Join(targetDir, "file3.txt")

	absSource1, err := filepath.Abs(sourceFile1)
	require.NoError(t, err)

	absSource2, err := filepath.Abs(sourceFile2)
	require.NoError(t, err)

	// Create valid symlinks for file1 and file2
	err = os.Symlink(absSource1, targetFile1)
	require.NoError(t, err)

	err = os.Symlink(absSource2, targetFile2)
	require.NoError(t, err)

	// Create a regular file (not symlink) for file3
	err = os.WriteFile(targetFile3, []byte("regular file"), 0644)
	require.NoError(t, err)

	t.Run("mixed scenario: valid symlinks and invalid targets", func(t *testing.T) {
		// Create state file with all three files
		stateFile := state.NewStateFile()
		stateFile.AddFileMapping(sourceFile1, targetFile1, state.TypeLink)
		stateFile.AddFileMapping(sourceFile2, targetFile2, state.TypeLink)
		stateFile.AddFileMapping(sourceFile3, targetFile3, state.TypeLink)

		statePath := filepath.Join(dotfilesDir, "state.yaml")
		err := state.SaveStateFile(statePath, stateFile)
		require.NoError(t, err)

		// Run uninstall
		result, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.RemovedLinks, 2) // file1 and file2 should be removed
		assert.Len(t, result.SkippedLinks, 1) // file3 should be skipped
		assert.Len(t, result.FailedRemovals, 0)
		assert.Len(t, result.Errors, 0)

		// Verify valid symlinks are removed
		assert.NoFileExists(t, targetFile1)
		assert.NoFileExists(t, targetFile2)

		// Verify regular file is not removed
		assert.FileExists(t, targetFile3)

		// Verify skipped reason
		assert.Contains(t, result.SkippedLinks[0].Reason, "target exists but is not a symlink")

		// Verify state file contains only the skipped file
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 1)
		assert.Equal(t, targetFile3, updatedStateFile.Files[0].Target)
	})
}

func TestUninstallNonLinkFiles(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")

	// Create directory
	err := os.MkdirAll(dotfilesDir, 0755)
	require.NoError(t, err)

	t.Run("skip non-link files in state file", func(t *testing.T) {
		// Create state file with generated file type
		stateFile := state.NewStateFile()
		stateFile.AddFileMapping("/source1", "/target1", state.TypeGenerated)
		stateFile.AddFileMapping("/source2", "/target2", state.TypeLink)

		statePath := filepath.Join(dotfilesDir, "state.yaml")
		err := state.SaveStateFile(statePath, stateFile)
		require.NoError(t, err)

		// Run uninstall
		result, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.RemovedLinks, 0) // No links should be removed
		assert.Len(t, result.SkippedLinks, 1) // Link file should be skipped (target doesn't exist)
		assert.Len(t, result.FailedRemovals, 0)

		// Verify state file contains both files (generated file is never processed, skipped link remains)
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 2) // Both files should remain
	})
}

package module

import (
	"crypto/sha1"
	"fmt"
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
		assert.Contains(t, result.Summary, "2 files removed (2 symlinks, 0 generated)")

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

		result := validateSymlink(fileMapping)
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

		result := validateSymlink(fileMapping)
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

		result := validateSymlink(fileMapping)
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

		result := validateSymlink(fileMapping)
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

		result := validateSymlink(fileMapping)
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
		err := removeSymlink(targetFile)
		assert.NoError(t, err)

		// Verify symlink is removed
		assert.NoFileExists(t, targetFile)
	})

	t.Run("remove non-existent symlink", func(t *testing.T) {
		err := removeSymlink(filepath.Join(tempDir, "nonexistent.txt"))
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
		assert.Len(t, result.RemovedLinks, 0)     // No links should be removed
		assert.Len(t, result.SkippedLinks, 1)     // Link file should be skipped (target doesn't exist)
		assert.Len(t, result.SkippedGenerated, 1) // Generated file should be skipped (target doesn't exist)
		assert.Len(t, result.FailedRemovals, 0)

		// Verify state file contains both files (both are skipped since targets don't exist)
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 2) // Both files should remain
	})
}

func TestValidateGeneratedFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid generated file with matching SHA1", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		content := "test content"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		// Calculate expected SHA1
		expectedSHA1 := calculateSHA1ForTest(content)

		fileMapping := state.FileMapping{
			Source: "/source",
			Target: targetFile,
			Type:   state.TypeGenerated,
			SHA1:   expectedSHA1,
		}

		result := validateGeneratedFile(fileMapping)
		assert.True(t, result.IsValid)
		assert.False(t, result.BackupRequired)
	})

	t.Run("generated file with mismatched SHA1", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		content := "modified content"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		fileMapping := state.FileMapping{
			Source: "/source",
			Target: targetFile,
			Type:   state.TypeGenerated,
			SHA1:   "different-sha1",
		}

		result := validateGeneratedFile(fileMapping)
		assert.True(t, result.IsValid) // Valid for removal but backup required
		assert.True(t, result.BackupRequired)
	})

	t.Run("generated file with empty SHA1 (backward compatibility)", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		content := "test content"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		fileMapping := state.FileMapping{
			Source: "/source",
			Target: targetFile,
			Type:   state.TypeGenerated,
			SHA1:   "", // Empty SHA1
		}

		result := validateGeneratedFile(fileMapping)
		assert.True(t, result.IsValid)
		assert.False(t, result.BackupRequired)
	})

	t.Run("target file does not exist", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "nonexistent.txt")

		fileMapping := state.FileMapping{
			Source: "/source",
			Target: targetFile,
			Type:   state.TypeGenerated,
			SHA1:   "some-sha1",
		}

		result := validateGeneratedFile(fileMapping)
		assert.False(t, result.IsValid)
		assert.Equal(t, "target file does not exist", result.Reason)
		assert.False(t, result.BackupRequired)
	})

	t.Run("target is not a regular file", func(t *testing.T) {
		targetDir := filepath.Join(tempDir, "testdir")
		err := os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		fileMapping := state.FileMapping{
			Source: "/source",
			Target: targetDir,
			Type:   state.TypeGenerated,
			SHA1:   "some-sha1",
		}

		result := validateGeneratedFile(fileMapping)
		assert.False(t, result.IsValid)
		assert.Equal(t, "target exists but is not a regular file", result.Reason)
		assert.False(t, result.BackupRequired)
	})
}

func TestCalculateSHA1(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("calculate SHA1 for file", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		content := "test content for SHA1"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		sha1, err := calculateSHA1(targetFile)
		require.NoError(t, err)
		assert.NotEmpty(t, sha1)

		// Verify it's the correct SHA1
		expectedSHA1 := calculateSHA1ForTest(content)
		assert.Equal(t, expectedSHA1, sha1)
	})

	t.Run("file does not exist", func(t *testing.T) {
		nonexistentFile := filepath.Join(tempDir, "nonexistent.txt")

		_, err := calculateSHA1(nonexistentFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open file")
	})
}

func TestCreateBackup(t *testing.T) {
	t.Run("create backup successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		targetFile := filepath.Join(tempDir, "test.txt")
		content := "test content"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		backupPath, err := createBackup(targetFile)
		require.NoError(t, err)
		assert.Equal(t, targetFile+".bak", backupPath)

		// Verify backup content
		backupContent, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, content, string(backupContent))
	})

	t.Run("handle existing backup file", func(t *testing.T) {
		tempDir := t.TempDir()
		targetFile := filepath.Join(tempDir, "test.txt")
		content := "test content"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		// Create first backup
		backupPath1, err := createBackup(targetFile)
		require.NoError(t, err)
		assert.Equal(t, targetFile+".bak", backupPath1)

		// Create second backup (should get different name)
		backupPath2, err := createBackup(targetFile)
		require.NoError(t, err)
		assert.Equal(t, targetFile+".bak.1", backupPath2)

		// Both backups should exist
		assert.FileExists(t, backupPath1)
		assert.FileExists(t, backupPath2)
	})

	t.Run("source file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		nonexistentFile := filepath.Join(tempDir, "nonexistent.txt")

		_, err := createBackup(nonexistentFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open source file")
	})
}

// Helper function to calculate SHA1 for test content
func calculateSHA1ForTest(content string) string {
	hasher := sha1.New()
	hasher.Write([]byte(content))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func TestUninstallWithGeneratedFiles(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	targetDir := filepath.Join(tempDir, "target")

	err := os.MkdirAll(dotfilesDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	t.Run("uninstall generated file with matching SHA1", func(t *testing.T) {
		// Create a generated file
		targetFile := filepath.Join(targetDir, "config.txt")
		content := "generated content"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		// Calculate SHA1
		sha1 := calculateSHA1ForTest(content)

		// Create state file with generated file entry
		stateFile := state.NewStateFile()
		stateFile.AddFileMapping("/dotfiles/config.txt", targetFile, state.TypeGenerated)
		// Manually set SHA1 since AddFileMapping calculates it for new files
		if len(stateFile.Files) > 0 {
			stateFile.Files[0].SHA1 = sha1
		}

		statePath := filepath.Join(dotfilesDir, "state.yaml")
		err = state.SaveStateFile(statePath, stateFile)
		require.NoError(t, err)

		// Run uninstall
		result, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.RemovedGenerated, 1)
		assert.Len(t, result.BackedUpGenerated, 0)
		assert.Len(t, result.SkippedGenerated, 0)

		// Verify file is removed
		assert.NoFileExists(t, targetFile)

		// Verify state file is updated
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 0)
	})

	t.Run("uninstall generated file with SHA1 mismatch creates backup", func(t *testing.T) {
		// Create a generated file
		targetFile := filepath.Join(targetDir, "modified.txt")
		originalContent := "original content"
		err := os.WriteFile(targetFile, []byte(originalContent), 0644)
		require.NoError(t, err)

		// Modify the file
		modifiedContent := "modified content"
		err = os.WriteFile(targetFile, []byte(modifiedContent), 0644)
		require.NoError(t, err)

		// Calculate SHA1 of original content
		originalSHA1 := calculateSHA1ForTest(originalContent)

		// Create state file with generated file entry
		stateFile := state.NewStateFile()
		stateFile.AddFileMapping("/dotfiles/modified.txt", targetFile, state.TypeGenerated)
		// Manually set SHA1 to original
		if len(stateFile.Files) > 0 {
			stateFile.Files[0].SHA1 = originalSHA1
		}

		statePath := filepath.Join(dotfilesDir, "state.yaml")
		err = state.SaveStateFile(statePath, stateFile)
		require.NoError(t, err)

		// Run uninstall
		result, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.RemovedGenerated, 0)
		assert.Len(t, result.BackedUpGenerated, 1)
		assert.Len(t, result.SkippedGenerated, 0)

		// Verify backup is created
		backupPath := targetFile + ".bak"
		assert.FileExists(t, backupPath)

		// Verify backup content is the modified content
		backupContent, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, modifiedContent, string(backupContent))

		// Verify original file still exists (modified files are not removed)
		assert.FileExists(t, targetFile)

		// Verify state file still contains the entry (backed up files remain)
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 1)
	})

	t.Run("uninstall mixed link and generated files", func(t *testing.T) {
		// Create a symlink
		sourceFile := filepath.Join(dotfilesDir, "link.txt")
		err := os.WriteFile(sourceFile, []byte("link content"), 0644)
		require.NoError(t, err)
		targetLink := filepath.Join(targetDir, "link.txt")
		err = os.Symlink(sourceFile, targetLink)
		require.NoError(t, err)

		// Create a generated file
		targetGenerated := filepath.Join(targetDir, "generated.txt")
		content := "generated content"
		err = os.WriteFile(targetGenerated, []byte(content), 0644)
		require.NoError(t, err)
		sha1 := calculateSHA1ForTest(content)

		// Create state file with both entries
		stateFile := state.NewStateFile()
		stateFile.AddFileMapping(sourceFile, targetLink, state.TypeLink)
		stateFile.AddFileMapping("/dotfiles/generated.txt", targetGenerated, state.TypeGenerated)
		// Set SHA1 for generated file
		for i := range stateFile.Files {
			if stateFile.Files[i].Type == state.TypeGenerated {
				stateFile.Files[i].SHA1 = sha1
			}
		}

		statePath := filepath.Join(dotfilesDir, "state.yaml")
		err = state.SaveStateFile(statePath, stateFile)
		require.NoError(t, err)

		// Run uninstall
		result, err := Uninstall(dotfilesDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.RemovedLinks, 1)
		assert.Len(t, result.RemovedGenerated, 1)
		assert.Len(t, result.BackedUpGenerated, 0)

		// Verify both files are removed
		assert.NoFileExists(t, targetLink)
		assert.NoFileExists(t, targetGenerated)

		// Verify state file is empty
		updatedStateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, updatedStateFile)
		assert.Len(t, updatedStateFile.Files, 0)
	})
}

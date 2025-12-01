package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymlinkManager_CreateSymlinkWithMkdir(t *testing.T) {
	tempDir := t.TempDir()
	fileOp := NewOperator()
	symlinkMgr := NewSymlinkManager(fileOp)

	t.Run("creates symlink when directory exists", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		err = symlinkMgr.CreateSymlinkWithMkdir(sourceFile, targetFile, false)
		require.NoError(t, err)

		assert.FileExists(t, targetFile)
		assert.True(t, fileOp.IsSymlink(targetFile))
	})

	t.Run("creates symlink and directory when mkdir=true", func(t *testing.T) {
		// Use a separate temp directory to ensure test isolation
		isolatedTempDir := t.TempDir()
		sourceFile := filepath.Join(isolatedTempDir, "source.txt")
		targetFile := filepath.Join(isolatedTempDir, "nested", "dir", "target.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		err = symlinkMgr.CreateSymlinkWithMkdir(sourceFile, targetFile, true)
		require.NoError(t, err)

		assert.FileExists(t, targetFile)
		assert.True(t, fileOp.IsSymlink(targetFile))
		assert.DirExists(t, filepath.Dir(targetFile))
	})

	t.Run("fails when directory doesn't exist and mkdir=false", func(t *testing.T) {
		// Use a separate temp directory to ensure test isolation
		isolatedTempDir := t.TempDir()
		sourceFile := filepath.Join(isolatedTempDir, "source.txt")
		targetFile := filepath.Join(isolatedTempDir, "nested", "dir", "target.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		err = symlinkMgr.CreateSymlinkWithMkdir(sourceFile, targetFile, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target directory does not exist")
		assert.NoFileExists(t, targetFile)
	})
}

func TestSymlinkManager_ValidateSymlink(t *testing.T) {
	fileOp := NewOperator()
	symlinkMgr := NewSymlinkManager(fileOp)

	t.Run("valid symlink returns true", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)
		err = os.Symlink(sourceFile, targetFile)
		require.NoError(t, err)

		isValid, reason, err := symlinkMgr.ValidateSymlink(targetFile, sourceFile)
		require.NoError(t, err)
		assert.True(t, isValid)
		assert.Empty(t, reason)
	})

	t.Run("non-existent target returns false", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "nonexistent.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		isValid, reason, err := symlinkMgr.ValidateSymlink(targetFile, sourceFile)
		require.NoError(t, err)
		assert.False(t, isValid)
		assert.Contains(t, reason, "target file does not exist")
	})

	t.Run("regular file returns false", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(targetFile, []byte("different content"), 0644)
		require.NoError(t, err)

		isValid, reason, err := symlinkMgr.ValidateSymlink(targetFile, sourceFile)
		require.NoError(t, err)
		assert.False(t, isValid)
		assert.Contains(t, reason, "target exists but is not a symlink")
	})

	t.Run("symlink pointing to wrong target returns false", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		wrongSource := filepath.Join(tempDir, "wrong.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(wrongSource, []byte("wrong content"), 0644)
		require.NoError(t, err)
		err = os.Symlink(wrongSource, targetFile)
		require.NoError(t, err)

		isValid, reason, err := symlinkMgr.ValidateSymlink(targetFile, sourceFile)
		require.NoError(t, err)
		assert.False(t, isValid)
		assert.Contains(t, reason, "symlink points to")
	})

	t.Run("relative symlink works correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		// Create relative symlink
		relSource := "source.txt"
		err = os.Symlink(relSource, targetFile)
		require.NoError(t, err)

		isValid, reason, err := symlinkMgr.ValidateSymlink(targetFile, sourceFile)
		require.NoError(t, err)
		assert.True(t, isValid)
		assert.Empty(t, reason)
	})
}

func TestSymlinkManager_RemoveSymlink(t *testing.T) {
	tempDir := t.TempDir()
	fileOp := NewOperator()
	symlinkMgr := NewSymlinkManager(fileOp)

	t.Run("removes existing symlink", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)
		err = os.Symlink(sourceFile, targetFile)
		require.NoError(t, err)

		err = symlinkMgr.RemoveSymlink(targetFile)
		assert.NoError(t, err)
		assert.NoFileExists(t, targetFile)
	})

	t.Run("fails on non-existent symlink", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		err := symlinkMgr.RemoveSymlink(nonExistentFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove symlink")
	})
}

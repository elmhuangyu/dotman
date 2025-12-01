package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupManager_CreateBackup(t *testing.T) {
	tempDir := t.TempDir()
	fileOp := NewOperator()
	backupMgr := NewBackupManager(fileOp)

	t.Run("creates backup successfully", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		content := "test content"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		backupPath, err := backupMgr.CreateBackup(targetFile)
		require.NoError(t, err)
		assert.Equal(t, targetFile+".bak", backupPath)

		// Verify backup content
		backupContent, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, content, string(backupContent))

		// Verify original file still exists
		assert.FileExists(t, targetFile)
	})

	t.Run("handles existing backup file", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test2.txt")
		content := "test content"
		err := os.WriteFile(targetFile, []byte(content), 0644)
		require.NoError(t, err)

		// Create first backup
		backupPath1, err := backupMgr.CreateBackup(targetFile)
		require.NoError(t, err)
		assert.Equal(t, targetFile+".bak", backupPath1)

		// Create second backup (should get different name)
		backupPath2, err := backupMgr.CreateBackup(targetFile)
		require.NoError(t, err)
		assert.Equal(t, targetFile+".bak.1", backupPath2)

		// Both backups should exist
		assert.FileExists(t, backupPath1)
		assert.FileExists(t, backupPath2)
	})

	t.Run("handles non-existing file", func(t *testing.T) {
		nonexistentFile := filepath.Join(tempDir, "nonexistent.txt")

		backupPath, err := backupMgr.CreateBackup(nonexistentFile)
		assert.Error(t, err)
		assert.Empty(t, backupPath)
	})
}

func TestBackupManager_BackupAndReplace(t *testing.T) {
	tempDir := t.TempDir()
	fileOp := NewOperator()
	backupMgr := NewBackupManager(fileOp)

	t.Run("replaces when no existing file", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		sourceFile := filepath.Join(tempDir, "source.txt")
		content := "source content"

		err := os.WriteFile(sourceFile, []byte(content), 0644)
		require.NoError(t, err)

		backupPath, err := backupMgr.BackupAndReplace(targetFile, func() error {
			return os.WriteFile(targetFile, []byte(content), 0644)
		})
		require.NoError(t, err)
		assert.Empty(t, backupPath)

		// Verify target file was created
		targetContent, err := os.ReadFile(targetFile)
		require.NoError(t, err)
		assert.Equal(t, content, string(targetContent))
	})

	t.Run("backs up and replaces existing file", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		sourceFile := filepath.Join(tempDir, "source.txt")
		originalContent := "original content"
		newContent := "new content"

		err := os.WriteFile(targetFile, []byte(originalContent), 0644)
		require.NoError(t, err)
		err = os.WriteFile(sourceFile, []byte(newContent), 0644)
		require.NoError(t, err)

		backupPath, err := backupMgr.BackupAndReplace(targetFile, func() error {
			return os.Symlink(sourceFile, targetFile)
		})
		require.NoError(t, err)
		assert.Equal(t, targetFile+".bak", backupPath)

		// Verify backup contains original content
		backupContent, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(backupContent))

		// Verify target is now a symlink
		assert.FileExists(t, targetFile)
		assert.True(t, fileOp.IsSymlink(targetFile))
	})

	t.Run("restores backup on replacement failure", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		originalContent := "original content"

		err := os.WriteFile(targetFile, []byte(originalContent), 0644)
		require.NoError(t, err)

		// Simulate replacement failure
		_, err = backupMgr.BackupAndReplace(targetFile, func() error {
			// First, let's backup succeed
			_, err := backupMgr.CreateBackup(targetFile)
			if err != nil {
				return err
			}
			// Then fail the replacement
			return os.WriteFile("/invalid/path/file", []byte("content"), 0644)
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "replacement failed")

		// Verify original content was restored
		restoredContent, err := os.ReadFile(targetFile)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(restoredContent))
	})

	t.Run("restores backup on replacement failure", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")
		originalContent := "original content"

		err := os.WriteFile(targetFile, []byte(originalContent), 0644)
		require.NoError(t, err)

		// Simulate replacement failure
		_, err = backupMgr.BackupAndReplace(targetFile, func() error {
			// First, let the backup succeed
			_, err := backupMgr.CreateBackup(targetFile)
			if err != nil {
				return err
			}
			// Then fail the replacement
			return os.WriteFile("/invalid/path/file", []byte("content"), 0644)
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "replacement failed")

		// Verify original content was restored
		restoredContent, err := os.ReadFile(targetFile)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(restoredContent))
	})
}

func TestBackupManager_ListBackups(t *testing.T) {
	tempDir := t.TempDir()
	fileOp := NewOperator()
	backupMgr := NewBackupManager(fileOp)

	t.Run("finds backup files", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "test.txt")

		// Create target file
		err := os.WriteFile(targetFile, []byte("content"), 0644)
		require.NoError(t, err)

		// Create multiple backups
		backup1, err := backupMgr.CreateBackup(targetFile)
		require.NoError(t, err)
		backup2, err := backupMgr.CreateBackup(targetFile)
		require.NoError(t, err)
		backup3, err := backupMgr.CreateBackup(targetFile)
		require.NoError(t, err)

		backups, err := backupMgr.ListBackups(targetFile)
		require.NoError(t, err)

		assert.Len(t, backups, 3)
		assert.Contains(t, backups, backup1)
		assert.Contains(t, backups, backup2)
		assert.Contains(t, backups, backup3)
	})

	t.Run("handles no backups", func(t *testing.T) {
		cleanTempDir := t.TempDir()
		targetFile := filepath.Join(cleanTempDir, "test.txt")

		backups, err := backupMgr.ListBackups(targetFile)
		require.NoError(t, err)
		assert.Empty(t, backups)
	})

	t.Run("filters unrelated files", func(t *testing.T) {
		// Use a separate temp directory to ensure test isolation
		isolatedTempDir := t.TempDir()
		targetFile := filepath.Join(isolatedTempDir, "test.txt")

		// Create target file
		err := os.WriteFile(targetFile, []byte("content"), 0644)
		require.NoError(t, err)

		// Create backup
		backup1, err := backupMgr.CreateBackup(targetFile)
		require.NoError(t, err)

		// Create unrelated files
		unrelatedFile := filepath.Join(isolatedTempDir, "other.txt.bak")
		err = os.WriteFile(unrelatedFile, []byte("unrelated"), 0644)
		require.NoError(t, err)

		backups, err := backupMgr.ListBackups(targetFile)
		require.NoError(t, err)

		assert.Len(t, backups, 1)
		assert.Contains(t, backups, backup1)
		assert.NotContains(t, backups, unrelatedFile)
	})
}

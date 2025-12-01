package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperator_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	op := NewOperator()

	t.Run("existing file returns true", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		assert.True(t, op.FileExists(testFile))
	})

	t.Run("non-existing file returns false", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		assert.False(t, op.FileExists(nonExistentFile))
	})

	t.Run("existing directory returns true", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := os.Mkdir(testDir, 0755)
		require.NoError(t, err)

		assert.True(t, op.FileExists(testDir))
	})
}

func TestOperator_IsSymlink(t *testing.T) {
	tempDir := t.TempDir()
	op := NewOperator()

	t.Run("regular file returns false", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		assert.False(t, op.IsSymlink(testFile))
	})

	t.Run("symlink returns true", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		sourceFile := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		err = os.Symlink(sourceFile, targetFile)
		require.NoError(t, err)

		assert.True(t, op.IsSymlink(targetFile))
	})

	t.Run("non-existing file returns false", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		assert.False(t, op.IsSymlink(nonExistentFile))
	})
}

func TestOperator_EnsureDirectory(t *testing.T) {
	tempDir := t.TempDir()
	op := NewOperator()

	t.Run("creates new directory", func(t *testing.T) {
		newDir := filepath.Join(tempDir, "new", "nested", "dir")

		err := op.EnsureDirectory(newDir)
		require.NoError(t, err)

		assert.DirExists(t, newDir)
	})

	t.Run("handles existing directory", func(t *testing.T) {
		existingDir := filepath.Join(tempDir, "existing")
		err := os.Mkdir(existingDir, 0755)
		require.NoError(t, err)

		err = op.EnsureDirectory(existingDir)
		assert.NoError(t, err)

		assert.DirExists(t, existingDir)
	})
}

func TestOperator_CreateSymlink(t *testing.T) {
	tempDir := t.TempDir()
	op := NewOperator()

	t.Run("creates symlink successfully", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		sourceFile := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		err = op.CreateSymlink(sourceFile, targetFile)
		require.NoError(t, err)

		assert.FileExists(t, targetFile)
		assert.True(t, op.IsSymlink(targetFile))

		linkTarget, err := op.Readlink(targetFile)
		require.NoError(t, err)
		assert.Equal(t, sourceFile, linkTarget)
	})
}

func TestOperator_RemoveFile(t *testing.T) {
	tempDir := t.TempDir()
	op := NewOperator()

	t.Run("removes existing file", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		err = op.RemoveFile(testFile)
		assert.NoError(t, err)
		assert.NoFileExists(t, testFile)
	})

	t.Run("removes existing symlink", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		sourceFile := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)
		err = os.Symlink(sourceFile, targetFile)
		require.NoError(t, err)

		err = op.RemoveFile(targetFile)
		assert.NoError(t, err)
		assert.NoFileExists(t, targetFile)
	})

	t.Run("handles non-existing file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		err := op.RemoveFile(nonExistentFile)
		assert.Error(t, err)
	})
}

func TestOperator_CopyFile(t *testing.T) {
	tempDir := t.TempDir()
	op := NewOperator()

	t.Run("copies file successfully", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		destFile := filepath.Join(tempDir, "dest.txt")
		content := "test content"

		err := os.WriteFile(sourceFile, []byte(content), 0644)
		require.NoError(t, err)

		err = op.CopyFile(sourceFile, destFile)
		require.NoError(t, err)

		assert.FileExists(t, destFile)

		destContent, err := os.ReadFile(destFile)
		require.NoError(t, err)
		assert.Equal(t, content, string(destContent))
	})

	t.Run("handles non-existing source", func(t *testing.T) {
		// Use a separate temp directory to ensure test isolation
		isolatedTempDir := t.TempDir()
		sourceFile := filepath.Join(isolatedTempDir, "nonexistent.txt")
		destFile := filepath.Join(isolatedTempDir, "dest.txt")

		err := op.CopyFile(sourceFile, destFile)
		assert.Error(t, err)
		assert.NoFileExists(t, destFile)
	})
}

func TestOperator_Readlink(t *testing.T) {
	tempDir := t.TempDir()
	op := NewOperator()

	t.Run("reads symlink target", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		sourceFile := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)
		err = os.Symlink(sourceFile, targetFile)
		require.NoError(t, err)

		linkTarget, err := op.Readlink(targetFile)
		require.NoError(t, err)
		assert.Equal(t, sourceFile, linkTarget)
	})

	t.Run("handles non-symlink", func(t *testing.T) {
		regularFile := filepath.Join(tempDir, "regular.txt")
		err := os.WriteFile(regularFile, []byte("content"), 0644)
		require.NoError(t, err)

		_, err = op.Readlink(regularFile)
		assert.Error(t, err)
	})

	t.Run("handles non-existing file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		_, err := op.Readlink(nonExistentFile)
		assert.Error(t, err)
	})
}

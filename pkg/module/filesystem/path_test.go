package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathResolver_EnsureDirExists(t *testing.T) {
	tempDir := t.TempDir()
	pr := NewPathResolver()

	t.Run("creates new directory", func(t *testing.T) {
		newDir := filepath.Join(tempDir, "new", "nested", "dir")

		err := pr.EnsureDirExists(newDir)
		require.NoError(t, err)

		assert.DirExists(t, newDir)
	})

	t.Run("handles existing directory", func(t *testing.T) {
		existingDir := filepath.Join(tempDir, "existing")
		err := os.Mkdir(existingDir, 0755)
		require.NoError(t, err)

		err = pr.EnsureDirExists(existingDir)
		assert.NoError(t, err)

		assert.DirExists(t, existingDir)
	})
}

func TestPathResolver_DirExists(t *testing.T) {
	tempDir := t.TempDir()
	pr := NewPathResolver()

	t.Run("existing directory returns true", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := os.Mkdir(testDir, 0755)
		require.NoError(t, err)

		assert.True(t, pr.DirExists(testDir))
	})

	t.Run("non-existing directory returns false", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		assert.False(t, pr.DirExists(nonExistentDir))
	})

	t.Run("regular file returns false", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		assert.False(t, pr.DirExists(testFile))
	})
}

func TestPathResolver_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	pr := NewPathResolver()

	t.Run("existing file returns true", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		assert.True(t, pr.FileExists(testFile))
	})

	t.Run("existing directory returns false", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := os.Mkdir(testDir, 0755)
		require.NoError(t, err)

		assert.False(t, pr.FileExists(testDir))
	})

	t.Run("non-existing file returns false", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		assert.False(t, pr.FileExists(nonExistentFile))
	})
}

package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadStateFile(t *testing.T) {
	t.Run("nonexistent file returns nil", func(t *testing.T) {
		stateFile, err := LoadStateFile("/nonexistent/path/state.yaml")
		assert.NoError(t, err)
		assert.Nil(t, stateFile)
	})

	t.Run("valid file loads correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := filepath.Join(tmpDir, "state.yaml")

		// Create a test state file
		testState := &StateFile{
			Version: "1.0.0",
			Files: []FileMapping{
				{
					Source: "/source/file1",
					Target: "/target/file1",
					Type:   TypeLink,
				},
				{
					Source: "/source/template1",
					Target: "/target/template1",
					Type:   TypeGenerated,
					SHA1:   "abc123def456",
				},
			},
		}

		err := SaveStateFile(statePath, testState)
		require.NoError(t, err)

		// Load the file
		loadedState, err := LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, loadedState)

		assert.Equal(t, testState.Version, loadedState.Version)
		assert.Len(t, loadedState.Files, 2)

		// Check link mapping
		assert.Equal(t, testState.Files[0].Source, loadedState.Files[0].Source)
		assert.Equal(t, testState.Files[0].Target, loadedState.Files[0].Target)
		assert.Equal(t, testState.Files[0].Type, loadedState.Files[0].Type)
		assert.Empty(t, loadedState.Files[0].SHA1)

		// Check generated mapping
		assert.Equal(t, testState.Files[1].Source, loadedState.Files[1].Source)
		assert.Equal(t, testState.Files[1].Target, loadedState.Files[1].Target)
		assert.Equal(t, testState.Files[1].Type, loadedState.Files[1].Type)
		assert.Equal(t, testState.Files[1].SHA1, loadedState.Files[1].SHA1)
	})

	t.Run("invalid file returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := filepath.Join(tmpDir, "state.yaml")

		// Create an invalid YAML file
		err := os.WriteFile(statePath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)

		// Try to load the file
		stateFile, err := LoadStateFile(statePath)
		assert.Error(t, err)
		assert.Nil(t, stateFile)
	})
}

func TestSaveStateFile(t *testing.T) {
	t.Run("saves state file correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := filepath.Join(tmpDir, "state.yaml")

		testState := &StateFile{
			Version: "1.0.0",
			Files: []FileMapping{
				{
					Source: "/source/file1",
					Target: "/target/file1",
					Type:   TypeLink,
				},
			},
		}

		err := SaveStateFile(statePath, testState)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(statePath)
		assert.NoError(t, err)

		// Load and verify content
		loadedState, err := LoadStateFile(statePath)
		require.NoError(t, err)
		assert.Equal(t, testState.Version, loadedState.Version)
		assert.Len(t, loadedState.Files, 1)
	})

	t.Run("creates directory if needed", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := filepath.Join(tmpDir, "nested", "dir", "state.yaml")

		testState := NewStateFile()

		err := SaveStateFile(statePath, testState)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(statePath)
		assert.NoError(t, err)
	})
}

func TestNewStateFile(t *testing.T) {
	stateFile := NewStateFile()

	assert.Equal(t, version, stateFile.Version)
	assert.NotNil(t, stateFile.Files)
	assert.Len(t, stateFile.Files, 0)
}

func TestCalculateSHA1(t *testing.T) {
	t.Run("calculates SHA1 for existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		// Create a test file with known content
		content := []byte("Hello, World!")
		err := os.WriteFile(testFile, content, 0644)
		require.NoError(t, err)

		// Calculate SHA1
		hash, err := calculateSHA1(testFile)
		require.NoError(t, err)

		// Expected SHA1 for "Hello, World!" (calculated manually)
		expected := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
		assert.Equal(t, expected, hash)
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		hash, err := calculateSHA1("/nonexistent/file")
		assert.Error(t, err)
		assert.Empty(t, hash)
	})
}

func TestAddFileMapping(t *testing.T) {
	t.Run("adds link mapping without SHA1", func(t *testing.T) {
		stateFile := NewStateFile()

		stateFile.AddFileMapping("/source/file1", "/target/file1", TypeLink)

		assert.Len(t, stateFile.Files, 1)
		assert.Equal(t, "/source/file1", stateFile.Files[0].Source)
		assert.Equal(t, "/target/file1", stateFile.Files[0].Target)
		assert.Equal(t, TypeLink, stateFile.Files[0].Type)
		assert.Empty(t, stateFile.Files[0].SHA1)
	})

	t.Run("adds generated mapping with SHA1", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "generated.txt")

		// Create a test file
		content := []byte("test content")
		err := os.WriteFile(testFile, content, 0644)
		require.NoError(t, err)

		stateFile := NewStateFile()
		stateFile.AddFileMapping("/source/template", testFile, TypeGenerated)

		assert.Len(t, stateFile.Files, 1)
		assert.Equal(t, "/source/template", stateFile.Files[0].Source)
		assert.Equal(t, testFile, stateFile.Files[0].Target)
		assert.Equal(t, TypeGenerated, stateFile.Files[0].Type)
		assert.NotEmpty(t, stateFile.Files[0].SHA1)
	})

	t.Run("handles SHA1 calculation error gracefully", func(t *testing.T) {
		stateFile := NewStateFile()

		// Try to add mapping for nonexistent file
		stateFile.AddFileMapping("/source/template", "/nonexistent/file", TypeGenerated)

		assert.Len(t, stateFile.Files, 1)
		assert.Equal(t, "/source/template", stateFile.Files[0].Source)
		assert.Equal(t, "/nonexistent/file", stateFile.Files[0].Target)
		assert.Equal(t, TypeGenerated, stateFile.Files[0].Type)
		assert.Empty(t, stateFile.Files[0].SHA1) // SHA1 should be empty on error
	})
}

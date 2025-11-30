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
			},
		}

		err := SaveStateFile(statePath, testState)
		require.NoError(t, err)

		// Load the file
		loadedState, err := LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, loadedState)

		assert.Equal(t, testState.Version, loadedState.Version)
		assert.Len(t, loadedState.Files, 1)
		assert.Equal(t, testState.Files[0].Source, loadedState.Files[0].Source)
		assert.Equal(t, testState.Files[0].Target, loadedState.Files[0].Target)
		assert.Equal(t, testState.Files[0].Type, loadedState.Files[0].Type)
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

func TestAddFileMapping(t *testing.T) {
	stateFile := NewStateFile()

	stateFile.AddFileMapping("/source/file1", "/target/file1", TypeLink)
	stateFile.AddFileMapping("/source/file2", "/target/file2", TypeGenerated)

	assert.Len(t, stateFile.Files, 2)

	// Check first mapping
	assert.Equal(t, "/source/file1", stateFile.Files[0].Source)
	assert.Equal(t, "/target/file1", stateFile.Files[0].Target)
	assert.Equal(t, TypeLink, stateFile.Files[0].Type)

	// Check second mapping
	assert.Equal(t, "/source/file2", stateFile.Files[1].Source)
	assert.Equal(t, "/target/file2", stateFile.Files[1].Target)
	assert.Equal(t, TypeGenerated, stateFile.Files[1].Type)
}

package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStateManager_Load(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.yaml")
	sm := NewStateManager()

	t.Run("nonexistent file returns nil", func(t *testing.T) {
		stateFile, err := sm.Load("/nonexistent/path/state.yaml")
		assert.NoError(t, err)
		assert.Nil(t, stateFile)
	})

	t.Run("valid file loads correctly", func(t *testing.T) {
		// Create a test state file
		testState := state.NewStateFile()
		testState.AddFileMapping("/source/file1", "/target/file1", state.TypeLink)
		testState.AddFileMapping("/source/template1", "/target/template1", state.TypeGenerated)

		err := sm.Save(statePath, testState)
		require.NoError(t, err)

		// Load the file
		loadedState, err := sm.Load(statePath)
		require.NoError(t, err)
		require.NotNil(t, loadedState)

		assert.Equal(t, testState.Version, loadedState.Version)
		assert.Len(t, loadedState.Files, 2)
	})

	t.Run("invalid file returns error", func(t *testing.T) {
		// Create an invalid YAML file
		invalidPath := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)

		// Try to load the file
		stateFile, err := sm.Load(invalidPath)
		assert.Error(t, err)
		assert.Nil(t, stateFile)
	})
}

func TestDefaultStateManager_Save(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.yaml")
	sm := NewStateManager()

	t.Run("saves state file correctly", func(t *testing.T) {
		testState := state.NewStateFile()
		testState.AddFileMapping("/source/file1", "/target/file1", state.TypeLink)

		err := sm.Save(statePath, testState)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(statePath)
		assert.NoError(t, err)

		// Load and verify content
		loadedState, err := sm.Load(statePath)
		require.NoError(t, err)
		assert.Equal(t, testState.Version, loadedState.Version)
		assert.Len(t, loadedState.Files, 1)
	})

	t.Run("creates directory if needed", func(t *testing.T) {
		nestedPath := filepath.Join(tmpDir, "nested", "dir", "state.yaml")
		testState := state.NewStateFile()

		err := sm.Save(nestedPath, testState)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(nestedPath)
		assert.NoError(t, err)
	})
}

func TestDefaultStateManager_AddMapping(t *testing.T) {
	sm := NewStateManager()
	stateFile := state.NewStateFile()

	t.Run("adds link mapping", func(t *testing.T) {
		err := sm.AddMapping(stateFile, "/source/file1", "/target/file1", state.TypeLink)
		require.NoError(t, err)

		assert.Len(t, stateFile.Files, 1)
		assert.Equal(t, "/source/file1", stateFile.Files[0].Source)
		assert.Equal(t, "/target/file1", stateFile.Files[0].Target)
		assert.Equal(t, state.TypeLink, stateFile.Files[0].Type)
	})

	t.Run("adds generated mapping", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "generated.txt")
		content := []byte("test content")
		err := os.WriteFile(testFile, content, 0644)
		require.NoError(t, err)

		err = sm.AddMapping(stateFile, "/source/template", testFile, state.TypeGenerated)
		require.NoError(t, err)

		assert.Len(t, stateFile.Files, 2)
		assert.Equal(t, "/source/template", stateFile.Files[1].Source)
		assert.Equal(t, testFile, stateFile.Files[1].Target)
		assert.Equal(t, state.TypeGenerated, stateFile.Files[1].Type)
		assert.NotEmpty(t, stateFile.Files[1].SHA1)
	})

	t.Run("prevents duplicate mappings", func(t *testing.T) {
		initialCount := len(stateFile.Files)

		// Try to add the same mapping again
		err := sm.AddMapping(stateFile, "/source/file1", "/target/file1", state.TypeLink)
		require.NoError(t, err)

		// Count should not have changed
		assert.Len(t, stateFile.Files, initialCount)
	})
}

func TestDefaultStateManager_RemoveMappings(t *testing.T) {
	sm := NewStateManager()
	stateFile := state.NewStateFile()

	// Add some test mappings
	stateFile.AddFileMapping("/source/file1", "/target/file1", state.TypeLink)
	stateFile.AddFileMapping("/source/file2", "/target/file2", state.TypeLink)
	stateFile.AddFileMapping("/source/file3", "/target/file3", state.TypeGenerated)

	t.Run("removes single mapping", func(t *testing.T) {
		targets := []string{"/target/file2"}
		err := sm.RemoveMappings(stateFile, targets)
		require.NoError(t, err)

		assert.Len(t, stateFile.Files, 2)
		assert.Equal(t, "/target/file1", stateFile.Files[0].Target)
		assert.Equal(t, "/target/file3", stateFile.Files[1].Target)
	})

	t.Run("removes multiple mappings", func(t *testing.T) {
		targets := []string{"/target/file1", "/target/file3"}
		err := sm.RemoveMappings(stateFile, targets)
		require.NoError(t, err)

		assert.Len(t, stateFile.Files, 0)
	})

	t.Run("handles empty target list", func(t *testing.T) {
		// Add a mapping first
		stateFile.AddFileMapping("/source/file4", "/target/file4", state.TypeLink)
		initialCount := len(stateFile.Files)

		targets := []string{}
		err := sm.RemoveMappings(stateFile, targets)
		require.NoError(t, err)

		// Count should not have changed
		assert.Len(t, stateFile.Files, initialCount)
	})

	t.Run("handles nonexistent targets", func(t *testing.T) {
		initialCount := len(stateFile.Files)

		targets := []string{"/nonexistent/target"}
		err := sm.RemoveMappings(stateFile, targets)
		require.NoError(t, err)

		// Count should not have changed
		assert.Len(t, stateFile.Files, initialCount)
	})
}

func TestNewStateManager(t *testing.T) {
	sm := NewStateManager()
	assert.NotNil(t, sm)
	assert.Implements(t, (*StateManager)(nil), sm)
}

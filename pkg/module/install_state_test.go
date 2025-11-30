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

func TestInstallWithStateLogging(t *testing.T) {
	t.Run("successful installation records state", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := filepath.Join(tmpDir, "state.yaml")

		// Override state path for testing
		originalHome := os.Getenv("HOME")
		t.Cleanup(func() {
			os.Setenv("HOME", originalHome)
		})
		os.Setenv("HOME", tmpDir)

		// Create test module
		sourceDir := filepath.Join(tmpDir, "source")
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create test file
		testFile := filepath.Join(sourceDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		module := config.ModuleConfig{
			Dir:       sourceDir,
			TargetDir: targetDir,
		}

		// Run install
		result, err := Install([]config.ModuleConfig{module}, true, false, tmpDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 1)

		// Check state file was created and contains the mapping
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile)

		assert.Equal(t, "1.0.0", stateFile.Version)
		assert.Len(t, stateFile.Files, 1)
		assert.Equal(t, testFile, stateFile.Files[0].Source)
		assert.Equal(t, filepath.Join(targetDir, "test.txt"), stateFile.Files[0].Target)
		assert.Equal(t, state.TypeLink, stateFile.Files[0].Type)
	})

	t.Run("installation with force mode records state", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := filepath.Join(tmpDir, "state.yaml")

		// Override state path for testing
		originalHome := os.Getenv("HOME")
		t.Cleanup(func() {
			os.Setenv("HOME", originalHome)
		})
		os.Setenv("HOME", tmpDir)

		// Create test module
		sourceDir := filepath.Join(tmpDir, "source")
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create test file
		testFile := filepath.Join(sourceDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Create existing target file
		targetFile := filepath.Join(targetDir, "test.txt")
		err = os.WriteFile(targetFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		module := config.ModuleConfig{
			Dir:       sourceDir,
			TargetDir: targetDir,
		}

		// Run install with force
		result, err := Install([]config.ModuleConfig{module}, true, true, tmpDir)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess)
		assert.Len(t, result.CreatedLinks, 1)

		// Check state file was created and contains the mapping
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		require.NotNil(t, stateFile)

		assert.Len(t, stateFile.Files, 1)
		assert.Equal(t, testFile, stateFile.Files[0].Source)
		assert.Equal(t, targetFile, stateFile.Files[0].Target)
		assert.Equal(t, state.TypeLink, stateFile.Files[0].Type)

		// Check backup file was created
		backupFile := targetFile + ".bak"
		_, err = os.Stat(backupFile)
		assert.NoError(t, err)
	})

	t.Run("failed installation does not record state for failed links", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := filepath.Join(tmpDir, "state.yaml")

		// Create test module with valid source but make target read-only to cause failure
		sourceDir := filepath.Join(tmpDir, "source")
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create test file
		testFile := filepath.Join(sourceDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Create target file with read-only permissions to cause symlink creation to fail
		targetFile := filepath.Join(targetDir, "test.txt")
		err = os.WriteFile(targetFile, []byte("existing content"), 0444) // read-only
		require.NoError(t, err)

		module := config.ModuleConfig{
			Dir:       sourceDir,
			TargetDir: targetDir,
		}

		// Run install without force (should fail due to conflict)
		result, err := Install([]config.ModuleConfig{module}, true, false, tmpDir)
		require.NoError(t, err)
		assert.False(t, result.IsSuccess)

		// State file should not exist or be empty
		stateFile, err := state.LoadStateFile(statePath)
		require.NoError(t, err)
		if stateFile != nil {
			assert.Len(t, stateFile.Files, 0)
		}
	})
}

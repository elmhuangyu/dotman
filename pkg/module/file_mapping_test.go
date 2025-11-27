package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileMapping(t *testing.T) {
	fm := NewFileMapping()
	assert.NotNil(t, fm)
	assert.NotNil(t, fm.sourceToTarget)
	assert.NotNil(t, fm.targetToSource)
	assert.Empty(t, fm.sourceToTarget)
	assert.Empty(t, fm.targetToSource)
}

func TestFileMappingAddMapping(t *testing.T) {
	fm := NewFileMapping()
	source := "/source/file.txt"
	target := "/target/file.txt"

	fm.AddMapping(source, target)

	// Test forward mapping
	retrievedTarget, exists := fm.GetTarget(source)
	assert.True(t, exists)
	assert.Equal(t, target, retrievedTarget)

	// Test reverse mapping
	retrievedSource, exists := fm.GetSource(target)
	assert.True(t, exists)
	assert.Equal(t, source, retrievedSource)
}

func TestFileMappingGetAllMappings(t *testing.T) {
	fm := NewFileMapping()
	mappings := map[string]string{
		"/source1.txt": "/target1.txt",
		"/source2.txt": "/target2.txt",
		"/source3.txt": "/target3.txt",
	}

	for source, target := range mappings {
		fm.AddMapping(source, target)
	}

	allMappings := fm.GetAllMappings()
	assert.Len(t, allMappings, 3)

	for source, target := range mappings {
		retrievedTarget, exists := allMappings[source]
		assert.True(t, exists)
		assert.Equal(t, target, retrievedTarget)
	}
}

func TestFileMappingGetTargetConflicts(t *testing.T) {
	fm := NewFileMapping()

	// Add mappings without conflicts
	fm.AddMapping("/source1.txt", "/target1.txt")
	fm.AddMapping("/source2.txt", "/target2.txt")
	conflicts := fm.GetTargetConflicts()
	assert.Empty(t, conflicts)

	// Add conflicting mappings
	fm.AddMapping("/source3.txt", "/target1.txt") // Conflict with source1
	fm.AddMapping("/source4.txt", "/target2.txt") // Conflict with source2

	conflicts = fm.GetTargetConflicts()
	assert.Len(t, conflicts, 2)

	// Check specific conflicts
	sources1, exists := conflicts["/target1.txt"]
	assert.True(t, exists)
	assert.Len(t, sources1, 2)
	assert.Contains(t, sources1, "/source1.txt")
	assert.Contains(t, sources1, "/source3.txt")

	sources2, exists := conflicts["/target2.txt"]
	assert.True(t, exists)
	assert.Len(t, sources2, 2)
	assert.Contains(t, sources2, "/source2.txt")
	assert.Contains(t, sources2, "/source4.txt")
}

func TestBuildFileMapping(t *testing.T) {
	tempDir := t.TempDir()

	// Create test module structure
	moduleDir := filepath.Join(tempDir, "test_module")
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "config.yaml"}
	for _, file := range testFiles {
		err := os.WriteFile(filepath.Join(moduleDir, file), []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Create Dotfile config
	dotfileContent := `target_dir: "/home/user/.config/test"
ignores:
  - "file1.txt"
`
	err = os.WriteFile(filepath.Join(moduleDir, "Dotfile"), []byte(dotfileContent), 0644)
	require.NoError(t, err)

	// Load module config
	moduleConfig, err := config.LoadConfig(moduleDir)
	require.NoError(t, err)
	require.NotNil(t, moduleConfig)

	// Build mapping
	mapping, err := BuildFileMapping([]config.ModuleConfig{*moduleConfig})
	require.NoError(t, err)
	require.NotNil(t, mapping)

	// Check that mapping exists and contains expected files
	allMappings := mapping.GetAllMappings()
	assert.NotEmpty(t, allMappings)

	// file1.txt should be ignored due to ignores list
	_, exists := mapping.GetTarget(filepath.Join(moduleDir, "file1.txt"))
	assert.False(t, exists)

	// file2.txt and config.yaml should be mapped
	_, exists = mapping.GetTarget(filepath.Join(moduleDir, "file2.txt"))
	assert.True(t, exists)

	_, exists = mapping.GetTarget(filepath.Join(moduleDir, "config.yaml"))
	assert.True(t, exists)
}

func TestBuildModuleMapping(t *testing.T) {
	tempDir := t.TempDir()

	// Create test module structure
	moduleDir := filepath.Join(tempDir, "test_module")
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "config.yaml"}
	for _, file := range testFiles {
		err := os.WriteFile(filepath.Join(moduleDir, file), []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Create Dotfile config
	dotfileContent := `target_dir: "/home/user/.config/test"
`
	err = os.WriteFile(filepath.Join(moduleDir, "Dotfile"), []byte(dotfileContent), 0644)
	require.NoError(t, err)

	// Load module config
	moduleConfig, err := config.LoadConfig(moduleDir)
	require.NoError(t, err)
	require.NotNil(t, moduleConfig)

	// Build mapping for single module
	mapping, err := buildModuleMapping(*moduleConfig)
	require.NoError(t, err)
	require.NotNil(t, mapping)

	// Check that all files (except Dotfile) are mapped
	allMappings := mapping.GetAllMappings()
	assert.Len(t, allMappings, 3) // 3 test files

	// Check specific mappings
	expectedTargets := map[string]string{
		filepath.Join(moduleDir, "file1.txt"):  "/home/user/.config/test/file1.txt",
		filepath.Join(moduleDir, "file2.txt"):  "/home/user/.config/test/file2.txt",
		filepath.Join(moduleDir, "config.yaml"): "/home/user/.config/test/config.yaml",
	}

	for source, expectedTarget := range expectedTargets {
		target, exists := mapping.GetTarget(source)
		assert.True(t, exists, "Source file %s should be mapped", source)
		assert.Equal(t, expectedTarget, target)
	}
}

func TestIsIgnored(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		ignores  []string
		expected bool
	}{
		{
			name:     "file not in ignores list",
			filename: "config.txt",
			ignores:  []string{"tmp", "cache"},
			expected: false,
		},
		{
			name:     "file matches ignore pattern",
			filename: "config.tmp",
			ignores:  []string{"tmp", "cache"},
			expected: true,
		},
		{
			name:     "file name exactly matches ignore",
			filename: "cache",
			ignores:  []string{"tmp", "cache"},
			expected: true,
		},
		{
			name:     "empty ignores list",
			filename: "any_file.txt",
			ignores:  []string{},
			expected: false,
		},
		{
			name:     "partial match should ignore",
			filename: "my_cache_file",
			ignores:  []string{"cache"},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isIgnored(test.filename, test.ignores)
			assert.Equal(t, test.expected, result)
		})
	}
}
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
	assert.NotNil(t, fm.templates)
	assert.Empty(t, fm.sourceToTarget)
	assert.Empty(t, fm.targetToSource)
	assert.Empty(t, fm.templates)
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

	// Test that regular files are not templates
	assert.False(t, fm.IsTemplate(source))
}

func TestFileMappingAddTemplateMapping(t *testing.T) {
	fm := NewFileMapping()
	source := "/source/config.dot-tmpl"
	target := "/target/config"

	fm.AddTemplateMapping(source, target)

	// Test forward mapping
	retrievedTarget, exists := fm.GetTarget(source)
	assert.True(t, exists)
	assert.Equal(t, target, retrievedTarget)

	// Test reverse mapping
	retrievedSource, exists := fm.GetSource(target)
	assert.True(t, exists)
	assert.Equal(t, source, retrievedSource)

	// Test that template is recognized
	assert.True(t, fm.IsTemplate(source))
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

func TestFileMappingGetTemplateMappings(t *testing.T) {
	fm := NewFileMapping()

	// Add regular mappings
	fm.AddMapping("/source1.txt", "/target1.txt")
	fm.AddMapping("/source2.txt", "/target2.txt")

	// Add template mappings
	fm.AddTemplateMapping("/source3.dot-tmpl", "/target3")
	fm.AddTemplateMapping("/source4.dot-tmpl", "/target4")

	templateMappings := fm.GetTemplateMappings()
	assert.Len(t, templateMappings, 2)

	expectedTemplates := map[string]string{
		"/source3.dot-tmpl": "/target3",
		"/source4.dot-tmpl": "/target4",
	}

	for source, target := range expectedTemplates {
		retrievedTarget, exists := templateMappings[source]
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
	testFiles := []string{"file1.txt", "file2.txt", "config.yaml", "template.dot-tmpl"}
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
	assert.Len(t, allMappings, 4) // 4 test files

	// Check specific mappings
	expectedTargets := map[string]string{
		filepath.Join(moduleDir, "file1.txt"):         "/home/user/.config/test/file1.txt",
		filepath.Join(moduleDir, "file2.txt"):         "/home/user/.config/test/file2.txt",
		filepath.Join(moduleDir, "config.yaml"):       "/home/user/.config/test/config.yaml",
		filepath.Join(moduleDir, "template.dot-tmpl"): "/home/user/.config/test/template",
	}

	for source, expectedTarget := range expectedTargets {
		target, exists := mapping.GetTarget(source)
		assert.True(t, exists, "Source file %s should be mapped", source)
		assert.Equal(t, expectedTarget, target)
	}

	// Check template detection
	templateSource := filepath.Join(moduleDir, "template.dot-tmpl")
	assert.True(t, mapping.IsTemplate(templateSource))

	// Check that regular files are not templates
	regularSource := filepath.Join(moduleDir, "file1.txt")
	assert.False(t, mapping.IsTemplate(regularSource))
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

func TestIsTemplateFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "regular file without extension",
			filename: "config",
			expected: false,
		},
		{
			name:     "regular file with .txt extension",
			filename: "config.txt",
			expected: false,
		},
		{
			name:     "template file with .dot-tmpl extension",
			filename: "config.dot-tmpl",
			expected: true,
		},
		{
			name:     "template file with name containing dots",
			filename: "my.config.dot-tmpl",
			expected: true,
		},
		{
			name:     "file ending with .tmpl but not .dot-tmpl",
			filename: "config.tmpl",
			expected: false,
		},
		{
			name:     "empty filename",
			filename: "",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isTemplateFile(test.filename)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestBuildModuleMappingWithSubdirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create test module structure with subdirectories
	moduleDir := filepath.Join(tempDir, "test_module")
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	// Create files in root directory
	err = os.WriteFile(filepath.Join(moduleDir, "root1.txt"), []byte("root content 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(moduleDir, "root2.txt"), []byte("root content 2"), 0644)
	require.NoError(t, err)

	// Create subdirectory with files
	subDir1 := filepath.Join(moduleDir, "subdir1")
	err = os.MkdirAll(subDir1, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir1, "sub1.txt"), []byte("sub content 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir1, "sub1.dot-tmpl"), []byte("template content 1"), 0644)
	require.NoError(t, err)

	// Create nested subdirectory
	nestedDir := filepath.Join(moduleDir, "subdir1", "nested")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(nestedDir, "nested.txt"), []byte("nested content"), 0644)
	require.NoError(t, err)

	// Create another subdirectory
	subDir2 := filepath.Join(moduleDir, "subdir2")
	err = os.MkdirAll(subDir2, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir2, "sub2.txt"), []byte("sub content 2"), 0644)
	require.NoError(t, err)

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

	// Check that all files (including subdirectories) are mapped
	allMappings := mapping.GetAllMappings()

	// We should have 6 files (root1.txt, root2.txt, sub1.txt, sub1.dot-tmpl, nested.txt, sub2.txt)
	// Dotfile should be excluded
	assert.Len(t, allMappings, 6, "Should map all files including subdirectories, excluding Dotfile")

	// Check specific mappings for root files
	root1Source := filepath.Join(moduleDir, "root1.txt")
	target1, exists := mapping.GetTarget(root1Source)
	assert.True(t, exists, "root1.txt should be mapped")
	assert.Equal(t, "/home/user/.config/test/root1.txt", target1)

	root2Source := filepath.Join(moduleDir, "root2.txt")
	target2, exists := mapping.GetTarget(root2Source)
	assert.True(t, exists, "root2.txt should be mapped")
	assert.Equal(t, "/home/user/.config/test/root2.txt", target2)

	// Check specific mappings for subdirectory files
	sub1Source := filepath.Join(moduleDir, "subdir1", "sub1.txt")
	target3, exists := mapping.GetTarget(sub1Source)
	assert.True(t, exists, "subdir1/sub1.txt should be mapped")
	assert.Equal(t, "/home/user/.config/test/subdir1/sub1.txt", target3)

	nestedSource := filepath.Join(moduleDir, "subdir1", "nested", "nested.txt")
	target4, exists := mapping.GetTarget(nestedSource)
	assert.True(t, exists, "subdir1/nested/nested.txt should be mapped")
	assert.Equal(t, "/home/user/.config/test/subdir1/nested/nested.txt", target4)

	sub2Source := filepath.Join(moduleDir, "subdir2", "sub2.txt")
	target5, exists := mapping.GetTarget(sub2Source)
	assert.True(t, exists, "subdir2/sub2.txt should be mapped")
	assert.Equal(t, "/home/user/.config/test/subdir2/sub2.txt", target5)

	// Check template file mapping
	templateSource := filepath.Join(moduleDir, "subdir1", "sub1.dot-tmpl")
	templateTarget, exists := mapping.GetTarget(templateSource)
	assert.True(t, exists, "subdir1/sub1.dot-tmpl should be mapped")
	assert.Equal(t, "/home/user/.config/test/subdir1/sub1", templateTarget) // .dot-tmpl extension removed

	// Check template detection
	assert.True(t, mapping.IsTemplate(templateSource), "Template file should be detected as template")

	// Check that regular files are not templates
	assert.False(t, mapping.IsTemplate(root1Source), "Regular file should not be detected as template")
}

func TestBuildModuleMappingWithSubdirectoriesAndIgnores(t *testing.T) {
	tempDir := t.TempDir()

	// Create test module structure with subdirectories
	moduleDir := filepath.Join(tempDir, "test_module")
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	// Create files in root directory
	err = os.WriteFile(filepath.Join(moduleDir, "root1.txt"), []byte("root content 1"), 0644)
	require.NoError(t, err)

	// Create subdirectory with files
	subDir1 := filepath.Join(moduleDir, "subdir1")
	err = os.MkdirAll(subDir1, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir1, "important.txt"), []byte("important content"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir1, "ignore_me.txt"), []byte("ignore content"), 0644)
	require.NoError(t, err)

	// Create another subdirectory that should be ignored
	ignoreDir := filepath.Join(moduleDir, "ignore_dir")
	err = os.MkdirAll(ignoreDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(ignoreDir, "file.txt"), []byte("should be ignored"), 0644)
	require.NoError(t, err)

	// Create Dotfile config with ignore patterns
	dotfileContent := `target_dir: "/home/user/.config/test"
ignores:
  - "ignore_me.txt"
  - "ignore_dir"
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

	// Check that only non-ignored files are mapped
	allMappings := mapping.GetAllMappings()

	// We should have 2 files: root1.txt and subdir1/important.txt
	// ignore_me.txt, ignore_dir/file.txt, and Dotfile should be excluded
	assert.Len(t, allMappings, 2, "Should only map non-ignored files")

	// Check that root1.txt is mapped
	root1Source := filepath.Join(moduleDir, "root1.txt")
	target1, exists := mapping.GetTarget(root1Source)
	assert.True(t, exists, "root1.txt should be mapped")
	assert.Equal(t, "/home/user/.config/test/root1.txt", target1)

	// Check that subdir1/important.txt is mapped
	importantSource := filepath.Join(moduleDir, "subdir1", "important.txt")
	target2, exists := mapping.GetTarget(importantSource)
	assert.True(t, exists, "subdir1/important.txt should be mapped")
	assert.Equal(t, "/home/user/.config/test/subdir1/important.txt", target2)

	// Check that ignored files are not mapped
	ignoreMeSource := filepath.Join(moduleDir, "subdir1", "ignore_me.txt")
	_, exists = mapping.GetTarget(ignoreMeSource)
	assert.False(t, exists, "ignore_me.txt should not be mapped")

	ignoreFileSource := filepath.Join(moduleDir, "ignore_dir", "file.txt")
	_, exists = mapping.GetTarget(ignoreFileSource)
	assert.False(t, exists, "ignore_dir/file.txt should not be mapped")
}

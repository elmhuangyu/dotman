package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_Render(t *testing.T) {
	tempDir := t.TempDir()
	renderer := NewRenderer()

	tests := []struct {
		name        string
		template    string
		vars        map[string]string
		expected    string
		expectError bool
	}{
		{
			name:     "simple variable substitution",
			template: "Hello {{.NAME}}!",
			vars:     map[string]string{"NAME": "World"},
			expected: "Hello World!",
		},
		{
			name:     "multiple variables",
			template: "User: {{.USER}}, Home: {{.HOME}}",
			vars:     map[string]string{"USER": "alice", "HOME": "/home/alice"},
			expected: "User: alice, Home: /home/alice",
		},
		{
			name:     "no variables",
			template: "Static content",
			vars:     map[string]string{},
			expected: "Static content",
		},
		{
			name:     "with ORIGINAL_FILE_PATH variable",
			template: "File: {{.ORIGINAL_FILE_PATH}}",
			vars:     map[string]string{},
			expected: "File: Original file: " + filepath.Join(tempDir, "test.tmpl"),
		},
		{
			name:        "missing variable",
			template:    "Hello {{.MISSING}}!",
			vars:        map[string]string{},
			expectError: true,
		},
		{
			name:        "invalid template syntax",
			template:    "Hello {{.NAME",
			vars:        map[string]string{"NAME": "World"},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create template file
			templatePath := filepath.Join(tempDir, "test.tmpl")
			err := os.WriteFile(templatePath, []byte(test.template), 0644)
			require.NoError(t, err)

			// Render template
			result, err := renderer.Render(templatePath, test.vars)

			if test.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, string(result))
			}
		})
	}
}

func TestRenderer_Validate(t *testing.T) {
	tempDir := t.TempDir()
	renderer := NewRenderer()

	tests := []struct {
		name        string
		template    string
		vars        map[string]string
		expectError bool
	}{
		{
			name:     "valid template with all variables",
			template: "Hello {{.NAME}}!",
			vars:     map[string]string{"NAME": "World"},
		},
		{
			name:     "valid template with no variables",
			template: "Static content",
			vars:     map[string]string{},
		},
		{
			name:     "valid template with ORIGINAL_FILE_PATH",
			template: "File: {{.ORIGINAL_FILE_PATH}}",
			vars:     map[string]string{},
		},
		{
			name:        "template with missing variable",
			template:    "Hello {{.MISSING}}!",
			vars:        map[string]string{},
			expectError: true,
		},
		{
			name:        "invalid template syntax",
			template:    "Hello {{.NAME",
			vars:        map[string]string{"NAME": "World"},
			expectError: true,
		},
		{
			name:        "template with undefined function",
			template:    "Hello {{.NAME | undefinedFunc}}!",
			vars:        map[string]string{"NAME": "World"},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create template file
			templatePath := filepath.Join(tempDir, "test.tmpl")
			err := os.WriteFile(templatePath, []byte(test.template), 0644)
			require.NoError(t, err)

			// Validate template
			err = renderer.Validate(templatePath, test.vars)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRenderer_RenderFileNotFound(t *testing.T) {
	renderer := NewRenderer()

	_, err := renderer.Render("/nonexistent/template.tmpl", map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read template file")
}

func TestRenderer_ValidateFileNotFound(t *testing.T) {
	renderer := NewRenderer()

	err := renderer.Validate("/nonexistent/template.tmpl", map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read template file")
}

func TestRenderer_VarsNotModified(t *testing.T) {
	tempDir := t.TempDir()
	renderer := NewRenderer()

	// Create template file
	templatePath := filepath.Join(tempDir, "test.tmpl")
	err := os.WriteFile(templatePath, []byte("Hello {{.NAME}}!"), 0644)
	require.NoError(t, err)

	// Original vars map
	vars := map[string]string{"NAME": "World"}
	originalVars := make(map[string]string)
	for k, v := range vars {
		originalVars[k] = v
	}

	// Render template
	_, err = renderer.Render(templatePath, vars)
	require.NoError(t, err)

	// Check that original vars map was not modified
	assert.Equal(t, originalVars, vars)
	assert.NotContains(t, vars, "ORIGINAL_FILE_PATH")
}

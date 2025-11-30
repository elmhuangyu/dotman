package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	tempDir := t.TempDir()

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
			result, err := RenderTemplate(templatePath, test.vars)

			if test.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, string(result))
			}
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	tempDir := t.TempDir()

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
			err = ValidateTemplate(templatePath, test.vars)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

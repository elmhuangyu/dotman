package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// Renderer implements TemplateRenderer interface
type Renderer struct{}

// NewRenderer creates a new template renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render renders a Go text template file using the provided variables
func (r *Renderer) Render(templatePath string, vars map[string]string) ([]byte, error) {
	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Get absolute path for ORIGINAL_FILE_PATH variable
	absPath, err := filepath.Abs(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", templatePath, err)
	}

	// Create a copy of vars to avoid modifying the original map
	templateVars := make(map[string]string)
	for k, v := range vars {
		templateVars[k] = v
	}
	templateVars["ORIGINAL_FILE_PATH"] = fmt.Sprintf("Original file: %s", absPath)

	// Parse the template with missingkey=error option
	tmpl, err := template.New("template").Option("missingkey=error").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Execute the template with variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateVars); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return buf.Bytes(), nil
}

// Validate validates a template file syntax and required variables
func (r *Renderer) Validate(templatePath string, vars map[string]string) error {
	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Get absolute path for ORIGINAL_FILE_PATH variable
	absPath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", templatePath, err)
	}

	// Create a copy of vars to avoid modifying the original map
	templateVars := make(map[string]string)
	for k, v := range vars {
		templateVars[k] = v
	}
	templateVars["ORIGINAL_FILE_PATH"] = fmt.Sprintf("Original file: %s", absPath)

	// Parse the template to check syntax
	tmpl, err := template.New("template").Option("missingkey=error").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("template syntax error in %s: %w", templatePath, err)
	}

	// Try to execute the template to check for missing variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateVars); err != nil {
		return fmt.Errorf("template execution error in %s: %w", templatePath, err)
	}

	return nil
}

package module

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// RenderTemplate renders a Go text template file using the provided variables
func RenderTemplate(templatePath string, vars map[string]string) ([]byte, error) {
	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Parse the template with missingkey=error option
	tmpl, err := template.New("template").Option("missingkey=error").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Execute the template with variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return buf.Bytes(), nil
}

// ValidateTemplate validates a template file syntax and required variables
func ValidateTemplate(templatePath string, vars map[string]string) error {
	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Parse the template to check syntax
	tmpl, err := template.New("template").Option("missingkey=error").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("template syntax error in %s: %w", templatePath, err)
	}

	// Try to execute the template to check for missing variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return fmt.Errorf("template execution error in %s: %w", templatePath, err)
	}

	return nil
}

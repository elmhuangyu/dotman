package template

// TemplateRenderer interface for template operations
type TemplateRenderer interface {
	Render(templatePath string, vars map[string]string) ([]byte, error)
	Validate(templatePath string, vars map[string]string) error
}

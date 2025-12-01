package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/module"
	"github.com/elmhuangyu/dotman/pkg/module/template"
)

// Validator handles validation of module installations
type Validator struct {
	templateRenderer template.TemplateRenderer
}

// NewValidator creates a new Validator instance
func NewValidator(templateRenderer template.TemplateRenderer) *Validator {
	return &Validator{
		templateRenderer: templateRenderer,
	}
}

// ValidateInstallation performs dry-run validation of the installation
func (v *Validator) ValidateInstallation(modules []config.ModuleConfig, vars map[string]string) (*ValidationResult, error) {
	// Build file mappings
	mapping, err := module.BuildFileMapping(modules)
	if err != nil {
		return nil, fmt.Errorf("failed to build file mappings: %v", err)
	}

	result := &ValidationResult{
		IsValid:  true,
		Mappings: mapping,
		Errors:   []string{},
	}

	// Check for target conflicts
	conflicts := mapping.GetTargetConflicts()
	for target, sources := range conflicts {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("target conflict: %d source files map to the same target %s: %v", len(sources), target, sources))
	}

	// Validate each mapping
	for source, target := range mapping.GetAllMappings() {
		operation, err := v.validateFileMapping(source, target, mapping.IsTemplate(source), vars)
		if err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("validation error for %s -> %s: %v", source, target, err))
			continue
		}

		result.Operations = append(result.Operations, operation)
	}

	return result, nil
}

// validateFileMapping validates a single source->target mapping
func (v *Validator) validateFileMapping(source, target string, isTemplate bool, vars map[string]string) (module.FileOperation, error) {
	// Check if source file exists
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return module.FileOperation{}, fmt.Errorf("source file does not exist: %s", source)
	}

	// Check source file info
	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return module.FileOperation{}, fmt.Errorf("failed to stat source file %s: %w", source, err)
	}

	if sourceInfo.IsDir() {
		return module.FileOperation{}, fmt.Errorf("source is a directory, not a file: %s", source)
	}

	// For templates, validate template syntax and variables
	if isTemplate {
		if err := v.templateRenderer.Validate(source, vars); err != nil {
			return module.FileOperation{}, fmt.Errorf("template validation failed: %w", err)
		}
	}

	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if os.IsNotExist(err) {
		// Target doesn't exist, this is a create operation
		if isTemplate {
			return module.FileOperation{
				Type:        module.OperationCreateTemplate,
				Source:      source,
				Target:      target,
				Description: "create new template file",
			}, nil
		} else {
			return module.FileOperation{
				Type:        module.OperationCreateLink,
				Source:      source,
				Target:      target,
				Description: "create new symlink",
			}, nil
		}
	} else if err != nil {
		return module.FileOperation{}, fmt.Errorf("failed to stat target %s: %w", target, err)
	}

	// For templates, we need to check if the target file exists and has correct content
	// For now, treat existing files as conflicts (will be handled by force mode)
	if isTemplate {
		return module.FileOperation{
			Type:        module.OperationForceTemplate,
			Source:      source,
			Target:      target,
			Description: "target exists as file (template would overwrite)",
		}, nil
	}

	// Target exists, check if it's a symlink to the correct source
	if targetInfo.Mode()&os.ModeSymlink != 0 {
		// Target is a symlink, check if it points to the correct source
		currentTarget, err := os.Readlink(target)
		if err != nil {
			return module.FileOperation{}, fmt.Errorf("failed to read symlink %s: %w", target, err)
		}

		// Resolve relative paths for comparison
		absSource, err := filepath.Abs(source)
		if err != nil {
			return module.FileOperation{}, fmt.Errorf("failed to resolve absolute path for source %s: %w", source, err)
		}

		absCurrentTarget, err := filepath.Abs(currentTarget)
		if err != nil {
			return module.FileOperation{}, fmt.Errorf("failed to resolve absolute path for current target %s: %w", currentTarget, err)
		}

		if absSource == absCurrentTarget {
			// Correct symlink already exists
			return module.FileOperation{
				Type:        module.OperationSkip,
				Source:      source,
				Target:      target,
				Description: "correct symlink already exists",
			}, nil
		} else {
			// Symlink exists but points to wrong file, treat as conflict
			return module.FileOperation{
				Type:        module.OperationForceLink,
				Source:      source,
				Target:      target,
				Description: fmt.Sprintf("target exists as symlink pointing to wrong file: %s", currentTarget),
			}, nil
		}
	} else {
		// Target exists but is not a symlink
		return module.FileOperation{
			Type:        module.OperationForceLink,
			Source:      source,
			Target:      target,
			Description: "target exists as regular file",
		}, nil
	}
}

// ValidateTargetDirectories ensures all target directories and their parents are valid
func (v *Validator) ValidateTargetDirectories(modules []config.ModuleConfig, mkdir bool) []string {
	var errors []string

	for _, module := range modules {
		// Validate target directory structure
		if err := v.validateDirectoryStructure(module.TargetDir, mkdir); err != nil {
			errors = append(errors, fmt.Sprintf("module %s: %v", module.Dir, err))
		}
	}

	return errors
}

// validateDirectoryStructure validates that a directory and all its parents are directories, not symlinks
func (v *Validator) validateDirectoryStructure(dir string, mkdir bool) error {
	// Start from the target directory and go up to root
	current := dir
	for {
		if current == "" || current == "/" || current == "." {
			break
		}

		// Check if path exists
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			// For the target directory itself, it must exist unless mkdir is enabled
			if current == dir {
				if mkdir {
					// With mkdir enabled, missing target directory is allowed
					break
				}
				return fmt.Errorf("target directory does not exist: %s", current)
			}
			// For parent directories, continue checking
			parent := filepath.Dir(current)
			if parent == current {
				break // We've reached the root
			}
			current = parent
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat %s: %w", current, err)
		}

		// Check if it's a symlink first
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path %s is a symlink, must be a regular directory", current)
		}

		// Check if it's a directory
		if !info.IsDir() {
			return fmt.Errorf("path %s exists but is not a directory", current)
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		if parent == current {
			break // We've reached the root
		}
		current = parent
	}

	return nil
}

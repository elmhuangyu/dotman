package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/module"
	"github.com/elmhuangyu/dotman/pkg/module/template"
)

// MappingValidator handles validation of file mappings
type MappingValidator struct {
	templateRenderer template.TemplateRenderer
}

// NewMappingValidator creates a new MappingValidator instance
func NewMappingValidator(templateRenderer template.TemplateRenderer) *MappingValidator {
	return &MappingValidator{
		templateRenderer: templateRenderer,
	}
}

// ValidateFileMapping validates a single source->target mapping
func (mv *MappingValidator) ValidateFileMapping(source, target string, isTemplate bool, vars map[string]string) (module.FileOperation, error) {
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
		if err := mv.templateRenderer.Validate(source, vars); err != nil {
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

// DetectConflicts detects target conflicts in file mappings
func (mv *MappingValidator) DetectConflicts(mapping *module.FileMapping) []ConflictError {
	var conflicts []ConflictError

	targetConflicts := mapping.GetTargetConflicts()
	for target, sources := range targetConflicts {
		conflicts = append(conflicts, ConflictError{
			Target:  target,
			Sources: sources,
			Message: fmt.Sprintf("target conflict: %d source files map to the same target %s: %v", len(sources), target, sources),
		})
	}

	return conflicts
}

// ConflictError represents a target conflict error
type ConflictError struct {
	Target  string
	Sources []string
	Message string
}

// Error implements the error interface
func (ce ConflictError) Error() string {
	return ce.Message
}

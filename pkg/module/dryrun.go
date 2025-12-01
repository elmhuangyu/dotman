package module

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/module/template"
)

// ValidateResult contains the complete results of a dry run
type ValidateResult struct {
	IsValid bool
	Summary string
	Errors  []string
	// Grouped operations by type
	CreateOperations    []FileOperation
	CreateTemplateOps   []FileOperation
	ForceLinkOperations []FileOperation
	ForceTemplateOps    []FileOperation
	SkipOperations      []FileOperation
}

// validateTargetDirectories ensures all target directories and their parents are valid
func validateTargetDirectories(modules []config.ModuleConfig, mkdir bool) []string {
	var errors []string

	for _, module := range modules {
		// Validate target directory structure
		if err := validateDirectoryStructure(module.TargetDir, mkdir); err != nil {
			errors = append(errors, fmt.Sprintf("module %s: %v", module.Dir, err))
		}
	}

	return errors
}

// validateDirectoryStructure validates that a directory and all its parents are directories, not symlinks
func validateDirectoryStructure(dir string, mkdir bool) error {
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

// validateFileMapping validates a single source->target mapping
func validateFileMapping(source, target string, isTemplate bool, vars map[string]string) (FileOperation, error) {
	// Check if source file exists
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return FileOperation{}, fmt.Errorf("source file does not exist: %s", source)
	}

	// Check source file info
	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return FileOperation{}, fmt.Errorf("failed to stat source file %s: %w", source, err)
	}

	if sourceInfo.IsDir() {
		return FileOperation{}, fmt.Errorf("source is a directory, not a file: %s", source)
	}

	// For templates, validate template syntax and variables
	if isTemplate {
		renderer := template.NewRenderer()
		if err := renderer.Validate(source, vars); err != nil {
			return FileOperation{}, fmt.Errorf("template validation failed: %w", err)
		}
	}

	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if os.IsNotExist(err) {
		// Target doesn't exist, this is a create operation
		if isTemplate {
			return FileOperation{
				Type:        OperationCreateTemplate,
				Source:      source,
				Target:      target,
				Description: "create new template file",
			}, nil
		} else {
			return FileOperation{
				Type:        OperationCreateLink,
				Source:      source,
				Target:      target,
				Description: "create new symlink",
			}, nil
		}
	} else if err != nil {
		return FileOperation{}, fmt.Errorf("failed to stat target %s: %w", target, err)
	}

	// For templates, we need to check if the target file exists and has correct content
	// For now, treat existing files as conflicts (will be handled by force mode)
	if isTemplate {
		return FileOperation{
			Type:        OperationForceTemplate,
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
			return FileOperation{}, fmt.Errorf("failed to read symlink %s: %w", target, err)
		}

		// Resolve relative paths for comparison
		absSource, err := filepath.Abs(source)
		if err != nil {
			return FileOperation{}, fmt.Errorf("failed to resolve absolute path for source %s: %w", source, err)
		}

		absCurrentTarget, err := filepath.Abs(currentTarget)
		if err != nil {
			return FileOperation{}, fmt.Errorf("failed to resolve absolute path for current target %s: %w", currentTarget, err)
		}

		if absSource == absCurrentTarget {
			// Correct symlink already exists
			return FileOperation{
				Type:        OperationSkip,
				Source:      source,
				Target:      target,
				Description: "correct symlink already exists",
			}, nil
		} else {
			// Symlink exists but points to wrong file, treat as conflict
			return FileOperation{
				Type:        OperationForceLink,
				Source:      source,
				Target:      target,
				Description: fmt.Sprintf("target exists as symlink pointing to wrong file: %s", currentTarget),
			}, nil
		}
	} else {
		// Target exists but is not a symlink
		return FileOperation{
			Type:        OperationForceLink,
			Source:      source,
			Target:      target,
			Description: "target exists as regular file",
		}, nil
	}
}

// validateInstallation performs dry-run validation of the installation
func validateInstallation(modules []config.ModuleConfig, vars map[string]string) (*struct {
	IsValid    bool
	Mappings   *FileMapping
	Errors     []string
	Operations []FileOperation
}, error) {
	// Build file mappings
	mapping, err := BuildFileMapping(modules)
	if err != nil {
		return nil, fmt.Errorf("failed to build file mappings: %v", err)
	}

	result := &struct {
		IsValid    bool
		Mappings   *FileMapping
		Errors     []string
		Operations []FileOperation
	}{
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
		operation, err := validateFileMapping(source, target, mapping.IsTemplate(source), vars)
		if err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("validation error for %s -> %s: %v", source, target, err))
			continue
		}

		result.Operations = append(result.Operations, operation)
	}

	return result, nil
}

// Validate performs a complete dry-run validation and returns structured results
func Validate(modules []config.ModuleConfig, vars map[string]string, mkdir bool, force bool) (*ValidateResult, error) {
	log := logger.GetLogger()

	log.Info().Int("modules", len(modules)).Msg("Starting validation")

	// Debug log all module names
	moduleNames := make([]string, len(modules))
	for i, module := range modules {
		_, moduleNames[i] = filepath.Split(module.Dir)
	}
	log.Debug().Str("modules", strings.Join(moduleNames, ", ")).Msg("Processing modules")

	// Validate target directories first
	dirErrors := validateTargetDirectories(modules, mkdir)
	if len(dirErrors) > 0 {
		return &ValidateResult{
			IsValid: false,
			Errors:  dirErrors,
		}, nil
	}

	// Validate file mappings
	validation, err := validateInstallation(modules, vars)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Group operations by type
	result := &ValidateResult{
		IsValid: validation.IsValid,
		Errors:  validation.Errors,
	}

	for _, op := range validation.Operations {
		switch op.Type {
		case OperationCreateLink:
			result.CreateOperations = append(result.CreateOperations, op)
		case OperationCreateTemplate:
			result.CreateTemplateOps = append(result.CreateTemplateOps, op)
		case OperationForceLink:
			result.ForceLinkOperations = append(result.ForceLinkOperations, op)
		case OperationForceTemplate:
			result.ForceTemplateOps = append(result.ForceTemplateOps, op)
		case OperationSkip:
			result.SkipOperations = append(result.SkipOperations, op)
		}
	}

	// Sort operations for consistent output
	sortFileOperations(result.CreateOperations)
	sortFileOperations(result.CreateTemplateOps)
	sortFileOperations(result.ForceLinkOperations)
	sortFileOperations(result.ForceTemplateOps)
	sortFileOperations(result.SkipOperations)

	// Force operations make the dry run invalid, unless in force mode
	// In force mode, only module config conflicts (multiple sources to same target) should fail
	// Target file conflicts (existing files) are allowed in force mode
	if (len(result.ForceLinkOperations) > 0 || len(result.ForceTemplateOps) > 0) && !force {
		result.IsValid = false
	}

	// Generate summary
	result.Summary = generateValidationSummary(result, force)

	log.Info().Bool("valid", result.IsValid).Msg("Validation completed")

	return result, nil
}

// sortFileOperations sorts operations by target path for consistent output
func sortFileOperations(ops []FileOperation) {
	sort.Slice(ops, func(i, j int) bool {
		return ops[i].Target < ops[j].Target
	})
}

// generateValidationSummary creates a human-readable summary of the validation results
func generateValidationSummary(result *ValidateResult, force bool) string {
	totalOps := len(result.CreateOperations) + len(result.CreateTemplateOps) + len(result.ForceLinkOperations) + len(result.ForceTemplateOps) + len(result.SkipOperations)

	summary := fmt.Sprintf("Validation Summary: %d total file operations\n", totalOps)

	if len(result.CreateOperations) > 0 {
		summary += fmt.Sprintf("  • %d files would be linked (new symlinks)\n", len(result.CreateOperations))
	}

	if len(result.CreateTemplateOps) > 0 {
		summary += fmt.Sprintf("  • %d template files would be generated\n", len(result.CreateTemplateOps))
	}

	forceOps := len(result.ForceLinkOperations) + len(result.ForceTemplateOps)
	if forceOps > 0 {
		if force {
			summary += fmt.Sprintf("  • %d conflicts found (will be backed up in force mode)\n", forceOps)
		} else {
			summary += fmt.Sprintf("  • %d conflicts found (targets exist as regular files or wrong symlinks)\n", forceOps)
		}
	}

	if len(result.SkipOperations) > 0 {
		summary += fmt.Sprintf("  • %d files skipped (correct symlinks already exist)\n", len(result.SkipOperations))
	}

	if len(result.Errors) > 0 {
		summary += fmt.Sprintf("  • %d errors\n", len(result.Errors))
	}

	return summary
}

// LogValidateResult logs the validation results in a structured format
func LogValidateResult(result *ValidateResult) {
	log := logger.GetLogger()

	// Log summary
	log.Info().Msg(result.Summary)

	// Log conflicts (these are the most important details)
	forceOps := append(result.ForceLinkOperations, result.ForceTemplateOps...)
	if len(forceOps) > 0 {
		log.Warn().Msg("Conflicts found:")
		for _, op := range forceOps {
			log.Warn().Msgf("  %s -> %s (%s)", op.Source, op.Target, op.Description)
		}
	}

	// Log errors
	if len(result.Errors) > 0 {
		log.Error().Msg("Errors:")
		for _, error := range result.Errors {
			log.Error().Msgf("  %s", error)
		}
	}
}

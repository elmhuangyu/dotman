package module

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
)

// ValidateResult contains the complete results of a dry run
type ValidateResult struct {
	IsValid bool
	Summary string
	Errors  []string
	// Grouped operations by type
	CreateOperations   []FileOperation
	SkipOperations     []FileOperation
	ConflictOperations []FileOperation
}

// Validate performs a complete dry-run validation and returns structured results
func Validate(modules []config.ModuleConfig, mkdir bool) (*ValidateResult, error) {
	log := logger.GetLogger()

	log.Info().Int("modules", len(modules)).Msg("Starting validation")

	// Debug log all module names
	moduleNames := make([]string, len(modules))
	for i, module := range modules {
		_, moduleNames[i] = filepath.Split(module.Dir)
	}
	log.Debug().Str("modules", strings.Join(moduleNames, ", ")).Msg("Processing modules")

	// Validate target directories first
	dirErrors := ValidateTargetDirectories(modules, mkdir)
	if len(dirErrors) > 0 {
		return &ValidateResult{
			IsValid: false,
			Errors:  dirErrors,
		}, nil
	}

	// Validate file mappings
	validation, err := ValidateInstallation(modules)
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
		case OperationSkip:
			result.SkipOperations = append(result.SkipOperations, op)
		case OperationConflict:
			result.ConflictOperations = append(result.ConflictOperations, op)
		}
	}

	// Sort operations for consistent output
	sortFileOperations(result.CreateOperations)
	sortFileOperations(result.SkipOperations)
	sortFileOperations(result.ConflictOperations)

	// Conflicts make the dry run invalid
	if len(result.ConflictOperations) > 0 {
		result.IsValid = false
	}

	// Generate summary
	result.Summary = generateValidationSummary(result)

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
func generateValidationSummary(result *ValidateResult) string {
	totalOps := len(result.CreateOperations) + len(result.SkipOperations) + len(result.ConflictOperations)

	summary := fmt.Sprintf("Validation Summary: %d total file operations\n", totalOps)

	if len(result.CreateOperations) > 0 {
		summary += fmt.Sprintf("  • %d files would be linked (new symlinks)\n", len(result.CreateOperations))
	}

	if len(result.SkipOperations) > 0 {
		summary += fmt.Sprintf("  • %d files skipped (correct symlinks already exist)\n", len(result.SkipOperations))
	}

	if len(result.ConflictOperations) > 0 {
		summary += fmt.Sprintf("  • %d conflicts found (targets exist as regular files or wrong symlinks)\n", len(result.ConflictOperations))
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
	if len(result.ConflictOperations) > 0 {
		log.Warn().Msg("Conflicts found:")
		for _, op := range result.ConflictOperations {
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

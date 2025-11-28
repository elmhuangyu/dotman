package module

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
)

// InstallResult contains the results of an installation
type InstallResult struct {
	IsSuccess    bool
	Summary      string
	Errors       []string
	CreatedLinks []FileOperation
	SkippedLinks []FileOperation
}

// Install performs the actual installation of dotfiles by creating symlinks
func Install(modules []config.ModuleConfig) (*InstallResult, error) {
	log := logger.GetLogger()

	log.Info().Int("modules", len(modules)).Msg("Starting installation")

	// First validate the installation
	validation, err := Validate(modules)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	result := &InstallResult{
		IsSuccess: true,
		Errors:    []string{},
	}

	// Check for validation errors or conflicts - if any exist, fail the installation
	if len(validation.Errors) > 0 {
		result.IsSuccess = false
		result.Errors = validation.Errors
		result.Summary = fmt.Sprintf("Installation failed: %d validation errors", len(validation.Errors))
		return result, nil
	}

	// Check for conflicts in the operations
	if len(validation.ConflictOperations) > 0 {
		result.IsSuccess = false
		result.Errors = append(result.Errors, "conflicts detected - installation would overwrite existing files")
		result.Summary = "Installation failed: conflicts detected"
		return result, nil
	}

	// Perform the installation
	for _, operation := range validation.CreateOperations {
		if err := createSymlink(operation.Source, operation.Target); err != nil {
			result.IsSuccess = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create symlink %s -> %s: %v", operation.Source, operation.Target, err))
		}
		result.CreatedLinks = append(result.CreatedLinks, operation)
		log.Info().Str("source", operation.Source).Str("target", operation.Target).Msg("Created symlink")

		if !result.IsSuccess {
			break
		}
	}

	// Generate summary
	if result.IsSuccess {
		result.Summary = fmt.Sprintf("Installation successful: %d symlinks created, %d skipped", len(result.CreatedLinks), len(result.SkippedLinks))
	} else {
		result.Summary = fmt.Sprintf("Installation failed: %d errors", len(result.Errors))
	}

	log.Info().Bool("success", result.IsSuccess).Msg("Installation completed")

	return result, nil
}

// createSymlink creates a symlink from source to target
func createSymlink(source, target string) error {
	// Ensure target directory exists
	targetDir := filepath.Dir(target)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("target directory does not exist: %s", targetDir)
	}

	// Get absolute path for source
	absSource, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source %s: %w", source, err)
	}

	// Create the symlink using absolute path
	if err := os.Symlink(absSource, target); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

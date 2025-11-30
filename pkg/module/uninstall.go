package module

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/state"
	"github.com/rs/zerolog"
)

// UninstallResult contains the results of an uninstallation
type UninstallResult struct {
	IsSuccess      bool
	Summary        string
	Errors         []string
	RemovedLinks   []FileOperation
	SkippedLinks   []OperationResult
	FailedRemovals []OperationResult
}

// OperationResult represents the result of a file operation with details
type OperationResult struct {
	Operation FileOperation
	Reason    string
}

// Uninstall performs the uninstallation of dotfiles using the state file
func Uninstall(dotfilesDir string) (*UninstallResult, error) {
	log := logger.GetLogger()

	log.Info().Str("dotfiles_dir", dotfilesDir).Msg("Starting uninstallation")

	// Load state file
	statePath := filepath.Join(dotfilesDir, "state.yaml")
	stateFile, err := state.LoadStateFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load state file: %w", err)
	}

	if stateFile == nil {
		log.Info().Msg("No state file found - no tracked installations to uninstall")
		return &UninstallResult{
			IsSuccess: true,
			Summary:   "No tracked installations found",
		}, nil
	}

	log.Info().Int("tracked_files", len(stateFile.Files)).Msg("Loaded state file")

	result := &UninstallResult{
		IsSuccess: true,
		Errors:    []string{},
	}

	// Process each file mapping in the state file
	for _, fileMapping := range stateFile.Files {
		if fileMapping.Type != state.TypeLink {
			log.Debug().Str("type", fileMapping.Type).Msg("Skipping non-link file type")
			continue
		}

		operation := FileOperation{
			Type:        OperationCreateLink, // Reuse this type for consistency
			Source:      fileMapping.Source,
			Target:      fileMapping.Target,
			Description: fmt.Sprintf("Remove symlink %s -> %s", fileMapping.Target, fileMapping.Source),
		}

		// Validate symlink before removal
		validationResult := validateSymlink(fileMapping)
		if !validationResult.IsValid {
			result.SkippedLinks = append(result.SkippedLinks, OperationResult{
				Operation: operation,
				Reason:    validationResult.Reason,
			})
			log.Warn().Str("target", fileMapping.Target).Str("reason", validationResult.Reason).Msg("Skipping symlink removal")
			continue
		}

		// Remove the symlink
		if err := removeSymlink(fileMapping.Target); err != nil {
			result.FailedRemovals = append(result.FailedRemovals, OperationResult{
				Operation: operation,
				Reason:    err.Error(),
			})
			result.Errors = append(result.Errors, fmt.Sprintf("failed to remove symlink %s: %v", fileMapping.Target, err))
			log.Error().Err(err).Str("target", fileMapping.Target).Msg("Failed to remove symlink")
			continue
		}

		result.RemovedLinks = append(result.RemovedLinks, operation)
		log.Info().Str("target", fileMapping.Target).Msg("Successfully removed symlink")
	}

	// Update state file to remove successfully uninstalled entries
	if len(result.RemovedLinks) > 0 {
		if err := updateStateFile(statePath, stateFile, result.RemovedLinks, log); err != nil {
			log.Warn().Err(err).Msg("Failed to update state file after uninstallation")
			// Don't fail the operation, but log the warning
		}
	}

	// Generate summary
	if result.IsSuccess {
		result.Summary = fmt.Sprintf("Uninstall successful: %d symlinks removed, %d skipped, %d failed",
			len(result.RemovedLinks), len(result.SkippedLinks), len(result.FailedRemovals))
	} else {
		result.Summary = fmt.Sprintf("Uninstall completed with errors: %d removed, %d skipped, %d failed",
			len(result.RemovedLinks), len(result.SkippedLinks), len(result.FailedRemovals))
	}

	log.Info().Bool("success", result.IsSuccess).
		Int("removed", len(result.RemovedLinks)).
		Int("skipped", len(result.SkippedLinks)).
		Int("failed", len(result.FailedRemovals)).
		Msg("Uninstall completed")

	return result, nil
}

// SymlinkValidationResult contains the result of symlink validation
type SymlinkValidationResult struct {
	IsValid bool
	Reason  string
}

// validateSymlink validates that a symlink points to the expected source
func validateSymlink(fileMapping state.FileMapping) SymlinkValidationResult {
	// Check if target exists
	targetInfo, err := os.Lstat(fileMapping.Target)
	if err != nil {
		if os.IsNotExist(err) {
			return SymlinkValidationResult{
				IsValid: false,
				Reason:  "target file does not exist",
			}
		}
		return SymlinkValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("failed to stat target: %v", err),
		}
	}

	// Check if target is a symlink
	if targetInfo.Mode()&os.ModeSymlink == 0 {
		return SymlinkValidationResult{
			IsValid: false,
			Reason:  "target exists but is not a symlink",
		}
	}

	// Read the symlink target
	actualSource, err := os.Readlink(fileMapping.Target)
	if err != nil {
		return SymlinkValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("failed to read symlink: %v", err),
		}
	}

	// Convert to absolute path for comparison
	if !filepath.IsAbs(actualSource) {
		// If relative, resolve relative to the symlink's directory
		actualSource = filepath.Join(filepath.Dir(fileMapping.Target), actualSource)
	}
	absActualSource, err := filepath.Abs(actualSource)
	if err != nil {
		return SymlinkValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("failed to resolve absolute path for actual source: %v", err),
		}
	}

	absExpectedSource, err := filepath.Abs(fileMapping.Source)
	if err != nil {
		return SymlinkValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("failed to resolve absolute path for expected source: %v", err),
		}
	}

	// Compare the paths
	if absActualSource != absExpectedSource {
		return SymlinkValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("symlink points to %s, expected %s", absActualSource, absExpectedSource),
		}
	}

	return SymlinkValidationResult{
		IsValid: true,
	}
}

// removeSymlink safely removes a symlink
func removeSymlink(target string) error {
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}
	return nil
}

// updateStateFile removes successfully uninstalled entries from the state file
func updateStateFile(statePath string, stateFile *state.StateFile, removedLinks []FileOperation, log zerolog.Logger) error {
	// Create a new list excluding the removed files
	var remainingFiles []state.FileMapping
	removedTargets := make(map[string]bool)

	// Build a set of removed targets for quick lookup
	for _, operation := range removedLinks {
		removedTargets[operation.Target] = true
	}

	// Filter out removed files
	for _, fileMapping := range stateFile.Files {
		if !removedTargets[fileMapping.Target] {
			remainingFiles = append(remainingFiles, fileMapping)
		}
	}

	// Update the state file
	stateFile.Files = remainingFiles

	// Save the updated state file
	if err := state.SaveStateFile(statePath, stateFile); err != nil {
		return fmt.Errorf("failed to save updated state file: %w", err)
	}

	log.Info().Int("remaining_files", len(remainingFiles)).Msg("Updated state file")
	return nil
}

package module

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/state"
	"github.com/rs/zerolog"
)

// UninstallResult contains the results of an uninstallation
type UninstallResult struct {
	IsSuccess         bool
	Summary           string
	Errors            []string
	RemovedLinks      []FileOperation
	SkippedLinks      []OperationResult
	RemovedGenerated  []FileOperation
	SkippedGenerated  []OperationResult
	BackedUpGenerated []OperationResult
	FailedRemovals    []OperationResult
}

// OperationResult represents the result of a file operation with details
type OperationResult struct {
	Operation FileOperation
	Reason    string
}

// Uninstall performs the uninstallation of dotfiles using the state file
func Uninstall(dotfilesDir string) (*UninstallResult, error) {
	log := logger.GetLogger()

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

	log.Debug().Int("tracked_files", len(stateFile.Files)).Msg("Loaded state file")

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
		log.Debug().Str("target", fileMapping.Target).Msg("Successfully removed symlink")
	}

	// Process generated files
	for _, fileMapping := range stateFile.Files {
		if fileMapping.Type != state.TypeGenerated {
			continue
		}

		operation := FileOperation{
			Type:        OperationCreateTemplate, // Reuse this type for consistency
			Source:      fileMapping.Source,
			Target:      fileMapping.Target,
			Description: fmt.Sprintf("Remove generated file %s", fileMapping.Target),
		}

		// Validate generated file before removal
		validationResult := validateGeneratedFile(fileMapping)
		if !validationResult.IsValid {
			result.SkippedGenerated = append(result.SkippedGenerated, OperationResult{
				Operation: operation,
				Reason:    validationResult.Reason,
			})
			log.Warn().Str("target", fileMapping.Target).Str("reason", validationResult.Reason).Msg("Skipping generated file removal")
			continue
		}

		// Check if file content has been modified
		if validationResult.BackupRequired {
			// Create backup and skip removal
			backupPath, err := createBackup(fileMapping.Target)
			if err != nil {
				result.FailedRemovals = append(result.FailedRemovals, OperationResult{
					Operation: operation,
					Reason:    fmt.Sprintf("failed to create backup: %v", err),
				})
				result.Errors = append(result.Errors, fmt.Sprintf("failed to backup generated file %s: %v", fileMapping.Target, err))
				log.Error().Err(err).Str("target", fileMapping.Target).Msg("Failed to create backup for modified generated file")
				continue
			}

			result.BackedUpGenerated = append(result.BackedUpGenerated, OperationResult{
				Operation: operation,
				Reason:    fmt.Sprintf("backed up to %s", backupPath),
			})
			log.Warn().Str("target", fileMapping.Target).Str("backup", backupPath).Msg("Created backup for modified generated file")
			// Skip removal for modified files - they remain in state for manual review
			continue
		}

		// Remove the generated file
		if err := removeGeneratedFile(fileMapping.Target); err != nil {
			result.FailedRemovals = append(result.FailedRemovals, OperationResult{
				Operation: operation,
				Reason:    err.Error(),
			})
			result.Errors = append(result.Errors, fmt.Sprintf("failed to remove generated file %s: %v", fileMapping.Target, err))
			log.Error().Err(err).Str("target", fileMapping.Target).Msg("Failed to remove generated file")
			continue
		}

		result.RemovedGenerated = append(result.RemovedGenerated, operation)
		log.Debug().Str("target", fileMapping.Target).Msg("Successfully removed generated file")
	}

	// Update state file to remove successfully uninstalled entries
	if len(result.RemovedLinks) > 0 || len(result.RemovedGenerated) > 0 {
		var allRemoved []FileOperation
		allRemoved = append(allRemoved, result.RemovedLinks...)
		allRemoved = append(allRemoved, result.RemovedGenerated...)
		if err := updateStateFile(statePath, stateFile, allRemoved, log); err != nil {
			log.Warn().Err(err).Msg("Failed to update state file after uninstallation")
			// Don't fail the operation, but log the warning
		}
	}

	// Generate summary
	totalRemoved := len(result.RemovedLinks) + len(result.RemovedGenerated)
	totalSkipped := len(result.SkippedLinks) + len(result.SkippedGenerated)
	if result.IsSuccess {
		result.Summary = fmt.Sprintf("Uninstall successful: %d files removed (%d symlinks, %d generated), %d skipped (%d symlinks, %d generated), %d backed up, %d failed",
			totalRemoved, len(result.RemovedLinks), len(result.RemovedGenerated),
			totalSkipped, len(result.SkippedLinks), len(result.SkippedGenerated),
			len(result.BackedUpGenerated), len(result.FailedRemovals))
	} else {
		result.Summary = fmt.Sprintf("Uninstall completed with errors: %d files removed (%d symlinks, %d generated), %d skipped (%d symlinks, %d generated), %d backed up, %d failed",
			totalRemoved, len(result.RemovedLinks), len(result.RemovedGenerated),
			totalSkipped, len(result.SkippedLinks), len(result.SkippedGenerated),
			len(result.BackedUpGenerated), len(result.FailedRemovals))
	}

	return result, nil
}

// SymlinkValidationResult contains the result of symlink validation
type SymlinkValidationResult struct {
	IsValid bool
	Reason  string
}

// GeneratedFileValidationResult contains the result of generated file validation
type GeneratedFileValidationResult struct {
	IsValid        bool
	Reason         string
	BackupRequired bool
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

// validateGeneratedFile validates a generated file for removal
func validateGeneratedFile(fileMapping state.FileMapping) GeneratedFileValidationResult {
	// Check if target exists
	targetInfo, err := os.Stat(fileMapping.Target)
	if err != nil {
		if os.IsNotExist(err) {
			return GeneratedFileValidationResult{
				IsValid:        false,
				Reason:         "target file does not exist",
				BackupRequired: false,
			}
		}
		return GeneratedFileValidationResult{
			IsValid:        false,
			Reason:         fmt.Sprintf("failed to stat target: %v", err),
			BackupRequired: false,
		}
	}

	// Check if target is a regular file
	if !targetInfo.Mode().IsRegular() {
		return GeneratedFileValidationResult{
			IsValid:        false,
			Reason:         "target exists but is not a regular file",
			BackupRequired: false,
		}
	}

	// Check SHA1 if available (for integrity verification)
	if fileMapping.SHA1 != "" {
		currentSHA1, err := calculateSHA1(fileMapping.Target)
		if err != nil {
			return GeneratedFileValidationResult{
				IsValid:        false,
				Reason:         fmt.Sprintf("failed to calculate SHA1: %v", err),
				BackupRequired: false,
			}
		}

		if currentSHA1 != fileMapping.SHA1 {
			return GeneratedFileValidationResult{
				IsValid:        true, // Valid for removal, but backup required
				Reason:         "file content has been modified",
				BackupRequired: true,
			}
		}
	}

	return GeneratedFileValidationResult{
		IsValid:        true,
		BackupRequired: false,
	}
}

// removeSymlink safely removes a symlink
func removeSymlink(target string) error {
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}
	return nil
}

// calculateSHA1 computes the SHA1 hash of a file's content
func calculateSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for SHA1 calculation: %w", err)
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file for SHA1 calculation: %w", err)
	}

	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}

// createBackup creates a backup of a file with .bak extension
func createBackup(target string) (string, error) {
	backupPath := target + ".bak"

	// Check if backup already exists and find a unique name if needed
	counter := 1
	for {
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			break // File doesn't exist, we can use this name
		}
		backupPath = fmt.Sprintf("%s.bak.%d", target, counter)
		counter++
		if counter > 100 { // Prevent infinite loop
			return "", fmt.Errorf("too many backup files exist")
		}
	}

	// Copy the file
	sourceFile, err := os.Open(target)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		os.Remove(backupPath) // Clean up on failure
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Ensure data is written
	if err := destFile.Sync(); err != nil {
		os.Remove(backupPath) // Clean up on failure
		return "", fmt.Errorf("failed to sync backup file: %w", err)
	}

	return backupPath, nil
}

// removeGeneratedFile safely removes a generated file
func removeGeneratedFile(target string) error {
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("failed to remove generated file: %w", err)
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

	log.Debug().Int("remaining_files", len(remainingFiles)).Msg("Updated state file")
	return nil
}

package module

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/module/filesystem"
	"github.com/elmhuangyu/dotman/pkg/module/state"
	dotmanState "github.com/elmhuangyu/dotman/pkg/state"
	"github.com/rs/zerolog"
)

// UninstallRequest contains the request parameters for uninstallation
type UninstallRequest struct {
	DotfilesDir    string
	BackupModified bool
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

// Uninstaller handles uninstallation operations with dependency injection
type Uninstaller struct {
	fileOp   filesystem.FileOperator
	stateMgr state.StateManager
}

// NewUninstaller creates a new Uninstaller instance
func NewUninstaller(fileOp filesystem.FileOperator, stateMgr state.StateManager) *Uninstaller {
	return &Uninstaller{
		fileOp:   fileOp,
		stateMgr: stateMgr,
	}
}

// Uninstall performs the uninstallation of dotfiles using the state file
func (u *Uninstaller) Uninstall(req *UninstallRequest) (*UninstallResult, error) {
	log := logger.GetLogger()

	// Load state file
	statePath := filepath.Join(req.DotfilesDir, "state.yaml")
	stateFile, err := u.stateMgr.Load(statePath)
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

	// Initialize filesystem operators
	symlinkMgr := filesystem.NewSymlinkManager(u.fileOp)
	backupMgr := filesystem.NewBackupManager(u.fileOp)

	// Process symlinks
	if err := u.uninstallSymlinks(stateFile, symlinkMgr, result); err != nil {
		return nil, fmt.Errorf("failed to uninstall symlinks: %w", err)
	}

	// Process generated files
	if err := u.uninstallGeneratedFiles(stateFile, backupMgr, result); err != nil {
		return nil, fmt.Errorf("failed to uninstall generated files: %w", err)
	}

	// Update state file to remove successfully uninstalled entries
	if err := u.updateStateFile(statePath, stateFile, result, log); err != nil {
		log.Warn().Err(err).Msg("Failed to update state file after uninstallation")
		// Don't fail the operation, but log the warning
	}

	// Generate summary
	u.generateSummary(result)

	return result, nil
}

// uninstallSymlinks processes all symlink mappings in the state file
func (u *Uninstaller) uninstallSymlinks(stateFile *dotmanState.StateFile, symlinkMgr *filesystem.SymlinkManager, result *UninstallResult) error {
	for _, fileMapping := range stateFile.Files {

		if fileMapping.Type != dotmanState.TypeLink {
			continue
		}

		operation := FileOperation{
			Type:        OperationCreateLink, // Reuse this type for consistency
			Source:      fileMapping.Source,
			Target:      fileMapping.Target,
			Description: fmt.Sprintf("Remove symlink %s -> %s", fileMapping.Target, fileMapping.Source),
		}

		// Validate symlink before removal
		if err := u.validateBeforeRemoval(fileMapping, symlinkMgr, result, operation); err != nil {
			continue // Skip this symlink, error already recorded
		}

		// Remove the symlink
		if err := u.removeSymlink(symlinkMgr, fileMapping.Target, result, operation); err != nil {
			continue // Error already recorded
		}

		result.RemovedLinks = append(result.RemovedLinks, operation)
		log := logger.GetLogger()
		log.Debug().Str("target", fileMapping.Target).Msg("Successfully removed symlink")
	}

	return nil
}

// uninstallGeneratedFiles processes all generated file mappings in the state file
func (u *Uninstaller) uninstallGeneratedFiles(stateFile *dotmanState.StateFile, backupMgr *filesystem.BackupManager, result *UninstallResult) error {
	for _, fileMapping := range stateFile.Files {

		if fileMapping.Type != dotmanState.TypeGenerated {
			continue
		}

		operation := FileOperation{
			Type:        OperationCreateTemplate, // Reuse this type for consistency
			Source:      fileMapping.Source,
			Target:      fileMapping.Target,
			Description: fmt.Sprintf("Remove generated file %s", fileMapping.Target),
		}

		// Validate generated file before removal
		validationResult := u.validateGeneratedFile(fileMapping)
		if !validationResult.IsValid {
			result.SkippedGenerated = append(result.SkippedGenerated, OperationResult{
				Type:     operation.Type,
				Source:   operation.Source,
				Target:   operation.Target,
				Success:  false,
				Error:    fmt.Errorf("validation failed: %s", validationResult.Reason),
				Metadata: map[string]interface{}{"reason": validationResult.Reason},
			})
			log := logger.GetLogger()
			log.Warn().Str("target", fileMapping.Target).Str("reason", validationResult.Reason).Msg("Skipping generated file removal")
			continue
		}

		// Check if file content has been modified and create backup if needed
		if validationResult.BackupRequired {
			if err := u.createBackupForGeneratedFile(backupMgr, fileMapping.Target, result, operation); err != nil {
				continue // Error already recorded
			}
		}

		// Remove the generated file
		if err := u.removeGeneratedFile(fileMapping.Target, result, operation); err != nil {
			continue // Error already recorded
		}

		result.RemovedGenerated = append(result.RemovedGenerated, operation)
		log := logger.GetLogger()
		log.Debug().Str("target", fileMapping.Target).Msg("Successfully removed generated file")
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

	// Use a buffered reader for reading
	buf := make([]byte, 32*1024) // 32KB buffer
	for {

		n, err := file.Read(buf)
		if n > 0 {
			if _, writeErr := hasher.Write(buf[:n]); writeErr != nil {
				return "", fmt.Errorf("failed to write to hasher: %w", writeErr)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to read file for SHA1 calculation: %w", err)
		}
	}

	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}

// validateBeforeRemoval validates a symlink before removal
func (u *Uninstaller) validateBeforeRemoval(fileMapping dotmanState.FileMapping, symlinkMgr *filesystem.SymlinkManager, result *UninstallResult, operation FileOperation) error {
	isValid, reason, err := symlinkMgr.ValidateSymlink(fileMapping.Target, fileMapping.Source)
	if err != nil {
		reason = fmt.Sprintf("failed to validate symlink: %v", err)
		isValid = false
	}

	if !isValid {
		result.SkippedLinks = append(result.SkippedLinks, OperationResult{
			Type:     operation.Type,
			Source:   operation.Source,
			Target:   operation.Target,
			Success:  false,
			Error:    fmt.Errorf("validation failed: %s", reason),
			Metadata: map[string]interface{}{"reason": reason},
		})
		log := logger.GetLogger()
		log.Warn().Str("target", fileMapping.Target).Str("reason", reason).Msg("Skipping symlink removal")
		return fmt.Errorf("validation failed: %s", reason)
	}
	return nil
}

// removeSymlink removes a symlink and records the result
func (u *Uninstaller) removeSymlink(symlinkMgr *filesystem.SymlinkManager, target string, result *UninstallResult, operation FileOperation) error {
	if err := symlinkMgr.RemoveSymlink(target); err != nil {
		result.FailedRemovals = append(result.FailedRemovals, OperationResult{
			Type:     operation.Type,
			Source:   operation.Source,
			Target:   operation.Target,
			Success:  false,
			Error:    err,
			Metadata: map[string]interface{}{"reason": err.Error()},
		})
		result.Errors = append(result.Errors, fmt.Sprintf("failed to remove symlink %s: %v", target, err))
		log := logger.GetLogger()
		log.Error().Err(err).Str("target", target).Msg("Failed to remove symlink")
		return err
	}
	return nil
}

// validateGeneratedFile validates a generated file for removal
func (u *Uninstaller) validateGeneratedFile(fileMapping dotmanState.FileMapping) GeneratedFileValidationResult {
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

// createBackupForGeneratedFile creates a backup for a modified generated file
func (u *Uninstaller) createBackupForGeneratedFile(backupMgr *filesystem.BackupManager, target string, result *UninstallResult, operation FileOperation) error {
	backupPath, err := backupMgr.CreateBackup(target)
	if err != nil {
		result.FailedRemovals = append(result.FailedRemovals, OperationResult{
			Type:     operation.Type,
			Source:   operation.Source,
			Target:   operation.Target,
			Success:  false,
			Error:    err,
			Metadata: map[string]interface{}{"reason": fmt.Sprintf("failed to create backup: %v", err)},
		})
		result.Errors = append(result.Errors, fmt.Sprintf("failed to backup generated file %s: %v", target, err))
		log := logger.GetLogger()
		log.Error().Err(err).Str("target", target).Msg("Failed to create backup for modified generated file")
		return err
	}

	result.BackedUpGenerated = append(result.BackedUpGenerated, OperationResult{
		Type:     operation.Type,
		Source:   operation.Source,
		Target:   operation.Target,
		Success:  true,
		Metadata: map[string]interface{}{"reason": fmt.Sprintf("backed up to %s", backupPath), "backup_path": backupPath},
	})
	log := logger.GetLogger()
	log.Warn().Str("target", target).Str("backup", backupPath).Msg("Created backup for modified generated file")
	return nil
}

// removeGeneratedFile removes a generated file and records the result
func (u *Uninstaller) removeGeneratedFile(target string, result *UninstallResult, operation FileOperation) error {
	if err := u.fileOp.RemoveFile(target); err != nil {
		result.FailedRemovals = append(result.FailedRemovals, OperationResult{
			Type:     operation.Type,
			Source:   operation.Source,
			Target:   operation.Target,
			Success:  false,
			Error:    err,
			Metadata: map[string]interface{}{"reason": err.Error()},
		})
		result.Errors = append(result.Errors, fmt.Sprintf("failed to remove generated file %s: %v", target, err))
		log := logger.GetLogger()
		log.Error().Err(err).Str("target", target).Msg("Failed to remove generated file")
		return err
	}
	return nil
}

// updateStateFile removes successfully uninstalled entries from the state file
func (u *Uninstaller) updateStateFile(statePath string, stateFile *dotmanState.StateFile, result *UninstallResult, log zerolog.Logger) error {
	if len(result.RemovedLinks) == 0 && len(result.RemovedGenerated) == 0 {
		return nil
	}

	// Collect all removed targets
	var removedTargets []string
	for _, op := range result.RemovedLinks {
		removedTargets = append(removedTargets, op.Target)
	}
	for _, op := range result.RemovedGenerated {
		removedTargets = append(removedTargets, op.Target)
	}

	// Remove mappings from state file
	if err := u.stateMgr.RemoveMappings(stateFile, removedTargets); err != nil {
		return fmt.Errorf("failed to remove mappings from state: %w", err)
	}

	// Save the updated state file
	if err := u.stateMgr.Save(statePath, stateFile); err != nil {
		return fmt.Errorf("failed to save updated state file: %w", err)
	}

	log.Debug().Int("remaining_files", len(stateFile.Files)).Msg("Updated state file")
	return nil
}

// generateSummary generates a summary of the uninstallation results
func (u *Uninstaller) generateSummary(result *UninstallResult) {
	totalRemoved := len(result.RemovedLinks) + len(result.RemovedGenerated)
	totalSkipped := len(result.SkippedLinks) + len(result.SkippedGenerated)

	if result.IsSuccess {
		result.Summary = fmt.Sprintf("Uninstall successful: %d files removed (%d symlinks, %d generated), %d skipped (%d symlinks, %d generated), %d backed up and removed, %d failed",
			totalRemoved, len(result.RemovedLinks), len(result.RemovedGenerated),
			totalSkipped, len(result.SkippedLinks), len(result.SkippedGenerated),
			len(result.BackedUpGenerated), len(result.FailedRemovals))
	} else {
		result.Summary = fmt.Sprintf("Uninstall completed with errors: %d files removed (%d symlinks, %d generated), %d skipped (%d symlinks, %d generated), %d backed up and removed, %d failed",
			totalRemoved, len(result.RemovedLinks), len(result.RemovedGenerated),
			totalSkipped, len(result.SkippedLinks), len(result.SkippedGenerated),
			len(result.BackedUpGenerated), len(result.FailedRemovals))
	}
}

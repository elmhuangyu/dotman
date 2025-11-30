package module

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/state"
	"github.com/rs/zerolog"
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

// Install performs the actual installation of dotfiles by creating symlinks
func Install(modules []config.ModuleConfig, mkdir bool, force bool, dotfilesDir string) (*InstallResult, error) {
	log := logger.GetLogger()

	log.Info().Int("modules", len(modules)).Msg("Starting installation")

	// Initialize state file
	var stateFile *state.StateFile
	var statePath string
	var err error

	if dotfilesDir != "" {
		statePath = filepath.Join(dotfilesDir, "state.yaml")
		stateFile, err = state.LoadStateFile(statePath)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load state file, continuing without state logging")
			stateFile = nil
		}
		if stateFile == nil {
			stateFile = state.NewStateFile()
		}
	}

	// First validate the installation
	validation, err := Validate(modules, mkdir, force)
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
	if len(validation.ConflictOperations) > 0 && !force {
		result.IsSuccess = false
		result.Errors = append(result.Errors, "conflicts detected - installation would overwrite existing files")
		result.Summary = "Installation failed: conflicts detected"
		return result, nil
	}

	result.SkippedLinks = validation.SkipOperations

	// Record skipped files in state file
	for _, operation := range validation.SkipOperations {
		if stateFile != nil {
			stateFile.AddFileMapping(operation.Source, operation.Target, state.TypeLink)
			if err := saveStateFile(statePath, stateFile, log); err != nil {
				log.Warn().Err(err).Msg("Failed to save state file for skipped operation")
			}
		}
		log.Info().Str("source", operation.Source).Str("target", operation.Target).Msg("Skipped (correct symlink already exists)")
	}

	// Perform the installation
	for _, operation := range validation.CreateOperations {
		if err := createSymlink(operation.Source, operation.Target, mkdir); err != nil {
			result.IsSuccess = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create symlink %s -> %s: %v", operation.Source, operation.Target, err))
		} else {
			// Record successful symlink in state file
			if stateFile != nil {
				stateFile.AddFileMapping(operation.Source, operation.Target, state.TypeLink)
				if err := saveStateFile(statePath, stateFile, log); err != nil {
					log.Warn().Err(err).Msg("Failed to save state file")
				}
			}
		}
		result.CreatedLinks = append(result.CreatedLinks, operation)
		log.Info().Str("source", operation.Source).Str("target", operation.Target).Msg("Created symlink")

		if !result.IsSuccess {
			break
		}
	}

	// Handle conflicts in force mode
	if force {
		for _, operation := range validation.ConflictOperations {
			if err := backupAndCreateSymlink(operation.Source, operation.Target, mkdir); err != nil {
				result.IsSuccess = false
				result.Errors = append(result.Errors, fmt.Sprintf("failed to backup and create symlink %s -> %s: %v", operation.Source, operation.Target, err))
			} else {
				// Record successful symlink in state file
				if stateFile != nil {
					stateFile.AddFileMapping(operation.Source, operation.Target, state.TypeLink)
					if err := saveStateFile(statePath, stateFile, log); err != nil {
						log.Warn().Err(err).Msg("Failed to save state file")
					}
				}
				result.CreatedLinks = append(result.CreatedLinks, operation)
				log.Warn().Str("source", operation.Source).Str("target", operation.Target).Msg("Backed up existing file and created symlink")
			}

			if !result.IsSuccess {
				break
			}
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

// saveStateFile saves the state file with error logging
func saveStateFile(path string, stateFile *state.StateFile, log zerolog.Logger) error {
	return state.SaveStateFile(path, stateFile)
}

// createSymlink creates a symlink from source to target
func createSymlink(source, target string, mkdir bool) error {
	// Ensure target directory exists
	targetDir := filepath.Dir(target)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if mkdir {
			// Create missing directories
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
			}
		} else {
			return fmt.Errorf("target directory does not exist: %s", targetDir)
		}
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

// backupAndCreateSymlink backs up the existing target file and creates a symlink
func backupAndCreateSymlink(source, target string, mkdir bool) error {
	// Ensure target directory exists
	targetDir := filepath.Dir(target)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if mkdir {
			// Create missing directories
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
			}
		} else {
			return fmt.Errorf("target directory does not exist: %s", targetDir)
		}
	}

	// Backup existing file
	backupPath := target + ".bak"
	if err := os.Rename(target, backupPath); err != nil {
		return fmt.Errorf("failed to backup existing file %s to %s: %w", target, backupPath, err)
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

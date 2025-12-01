package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/config"
)

// DirectoryValidator handles validation of directory structures
type DirectoryValidator struct{}

// NewDirectoryValidator creates a new DirectoryValidator instance
func NewDirectoryValidator() *DirectoryValidator {
	return &DirectoryValidator{}
}

// ValidateTargetDirectories ensures all target directories and their parents are valid
func (dv *DirectoryValidator) ValidateTargetDirectories(modules []config.ModuleConfig, mkdir bool) []DirectoryError {
	var errors []DirectoryError

	for _, module := range modules {
		// Validate target directory structure
		if err := dv.validateDirectoryStructure(module.TargetDir, mkdir); err != nil {
			errors = append(errors, DirectoryError{
				Module:  module.Dir,
				Path:    module.TargetDir,
				Message: err.Error(),
			})
		}
	}

	return errors
}

// validateDirectoryStructure validates that a directory and all its parents are directories, not symlinks
func (dv *DirectoryValidator) validateDirectoryStructure(dir string, mkdir bool) error {
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

// ValidateDirectory validates a single directory
func (dv *DirectoryValidator) ValidateDirectory(path string, mkdir bool) error {
	return dv.validateDirectoryStructure(path, mkdir)
}

// DirectoryError represents a directory validation error
type DirectoryError struct {
	Module  string
	Path    string
	Message string
}

// Error implements the error interface
func (de DirectoryError) Error() string {
	return de.Message
}

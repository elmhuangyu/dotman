package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// SymlinkManager handles symlink operations
type SymlinkManager struct {
	fileOp FileOperator
}

// NewSymlinkManager creates a new SymlinkManager
func NewSymlinkManager(fileOp FileOperator) *SymlinkManager {
	return &SymlinkManager{fileOp: fileOp}
}

// CreateSymlinkWithMkdir creates a symlink, ensuring the target directory exists
func (sm *SymlinkManager) CreateSymlinkWithMkdir(source, target string, mkdir bool) error {
	// Ensure target directory exists
	targetDir := filepath.Dir(target)
	if !sm.fileOp.FileExists(targetDir) {
		if mkdir {
			if err := sm.fileOp.EnsureDirectory(targetDir); err != nil {
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
	if err := sm.fileOp.CreateSymlink(absSource, target); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// ValidateSymlink validates that a symlink points to the expected source
func (sm *SymlinkManager) ValidateSymlink(target, expectedSource string) (bool, string, error) {
	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "target file does not exist", nil
		}
		return false, "", fmt.Errorf("failed to stat target: %w", err)
	}

	// Check if target is a symlink
	if targetInfo.Mode()&os.ModeSymlink == 0 {
		return false, "target exists but is not a symlink", nil
	}

	// Read the symlink target
	actualSource, err := sm.fileOp.Readlink(target)
	if err != nil {
		return false, "", fmt.Errorf("failed to read symlink: %w", err)
	}

	// Convert to absolute path for comparison
	if !filepath.IsAbs(actualSource) {
		// If relative, resolve relative to the symlink's directory
		actualSource = filepath.Join(filepath.Dir(target), actualSource)
	}
	absActualSource, err := filepath.Abs(actualSource)
	if err != nil {
		return false, "", fmt.Errorf("failed to resolve absolute path for actual source: %w", err)
	}

	absExpectedSource, err := filepath.Abs(expectedSource)
	if err != nil {
		return false, "", fmt.Errorf("failed to resolve absolute path for expected source: %w", err)
	}

	// Compare the paths
	if absActualSource != absExpectedSource {
		return false, fmt.Sprintf("symlink points to %s, expected %s", absActualSource, absExpectedSource), nil
	}

	return true, "", nil
}

// RemoveSymlink safely removes a symlink
func (sm *SymlinkManager) RemoveSymlink(target string) error {
	if err := sm.fileOp.RemoveFile(target); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}
	return nil
}

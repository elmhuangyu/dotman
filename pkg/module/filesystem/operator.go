package filesystem

import (
	"fmt"
	"io"
	"os"
)

// FileOperator interface for file system operations
type FileOperator interface {
	CreateSymlink(source, target string) error
	RemoveFile(path string) error
	CreateBackup(path string) (string, error)
	EnsureDirectory(path string) error
	CopyFile(src, dst string) error
	FileExists(path string) bool
	IsSymlink(path string) bool
	Readlink(path string) (string, error)
}

// Operator implements the FileOperator interface
type Operator struct{}

// NewOperator creates a new FileOperator instance
func NewOperator() FileOperator {
	return &Operator{}
}

// CreateSymlink creates a symlink from source to target
func (op *Operator) CreateSymlink(source, target string) error {
	return os.Symlink(source, target)
}

// RemoveFile removes a file or symlink
func (op *Operator) RemoveFile(path string) error {
	return os.Remove(path)
}

// EnsureDirectory ensures that a directory exists, creating it if necessary
func (op *Operator) EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// CopyFile copies a file from src to dst
func (op *Operator) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		os.Remove(dst) // Clean up on failure
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Ensure data is written
	if err := destFile.Sync(); err != nil {
		os.Remove(dst) // Clean up on failure
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists
func (op *Operator) FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsSymlink checks if a path is a symlink
func (op *Operator) IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// Readlink reads the target of a symlink
func (op *Operator) Readlink(path string) (string, error) {
	return os.Readlink(path)
}

// CreateBackup creates a backup of a file with .bak extension
func (op *Operator) CreateBackup(target string) (string, error) {
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
	if err := op.CopyFile(target, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

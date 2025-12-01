package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// BackupManager handles backup operations
type BackupManager struct {
	fileOp FileOperator
}

// NewBackupManager creates a new BackupManager
func NewBackupManager(fileOp FileOperator) *BackupManager {
	return &BackupManager{fileOp: fileOp}
}

// CreateBackup creates a backup of a file with .bak extension
func (bm *BackupManager) CreateBackup(target string) (string, error) {
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
	if err := bm.fileOp.CopyFile(target, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// BackupAndReplace backs up an existing file and replaces it with a new operation
func (bm *BackupManager) BackupAndReplace(target string, replaceFunc func() error) (string, error) {
	if !bm.fileOp.FileExists(target) {
		// No existing file, just perform the replacement
		if err := replaceFunc(); err != nil {
			return "", fmt.Errorf("failed to replace file: %w", err)
		}
		return "", nil
	}

	// Create backup by moving the existing file
	backupPath, err := bm.createBackupByMoving(target)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Perform the replacement
	if err := replaceFunc(); err != nil {
		// If replacement fails, try to restore from backup
		if restoreErr := os.Rename(backupPath, target); restoreErr != nil {
			return "", fmt.Errorf("replacement failed (%v) and restore from backup failed (%v)", err, restoreErr)
		}
		return "", fmt.Errorf("replacement failed, restored from backup: %w", err)
	}

	return backupPath, nil
}

// createBackupByMove creates a backup by moving the existing file (original behavior)
func (bm *BackupManager) createBackupByMoving(target string) (string, error) {
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

	// Move the file to backup location
	if err := os.Rename(target, backupPath); err != nil {
		return "", fmt.Errorf("failed to move file to backup: %w", err)
	}

	return backupPath, nil
}

// ListBackups finds all backup files for a given target
func (bm *BackupManager) ListBackups(target string) ([]string, error) {
	dir := filepath.Dir(target)
	base := filepath.Base(target)

	var backups []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Check if it's a backup of the target file
		if name == base+".bak" || (len(name) > len(base)+5 && name[:len(base)+5] == base+".bak.") {
			backups = append(backups, filepath.Join(dir, name))
		}
	}

	return backups, nil
}

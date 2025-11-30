# Change: Implement Uninstall with State File

## Why
To enable safe and tracked uninstallation of dotfiles using the state file that records all installed files. This ensures that only files that were actually installed by dotman are removed, and validates that symlinks are correct before deletion to prevent accidental removal of user files.

## What Changes
- Implement uninstall logic using state file to track which files to remove
- Add symlink validation before deletion to ensure correct target
- Skip removal with warning when current symlink doesn't match state file
- Update state file after successful uninstallation
- Add error handling for cases where state file doesn't exist or is corrupted

## Impact
- Affected specs: install (MODIFIED), uninstall (NEW)
- Affected code: cmd/uninstall.go, pkg/module/ (NEW uninstall.go), pkg/state/state_file.go
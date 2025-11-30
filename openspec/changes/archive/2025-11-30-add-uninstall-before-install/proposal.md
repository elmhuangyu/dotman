# Add Uninstall Before Install

## Summary
Modify the install command to automatically run uninstall before installing to ensure a clean state and prevent conflicts from previous installations.

## Problem
Currently, the install command does not clean up previous installations before creating new symlinks. This can lead to:
- Stale symlinks pointing to old sources
- Conflicts when module configurations change
- Inconsistent state between the filesystem and state file
- Manual cleanup requirements when updating dotfile configurations

## Solution
Make the install command automatically call uninstall before proceeding with installation. This ensures:
- Clean state before each installation
- Removal of obsolete or conflicting symlinks
- Consistent state file management
- Better user experience with automatic cleanup

## Scope
- Affected specs: install (MODIFIED), uninstall (MODIFIED)
- Affected code: cmd/install.go, pkg/module/install.go
- New behavior: install command runs uninstall first (except in dry-run mode)
- Backward compatibility: maintained (install still works the same, just with cleanup)

## Design Considerations
- Uninstall should only run when not in dry-run mode
- Uninstall failures should not prevent installation from proceeding
- Clear logging to show uninstall progress before install
- Preserve existing force and mkdir flag behaviors
- Handle cases where no state file exists gracefully

## Alternatives Considered
1. Manual cleanup requirement - rejected for poor UX
2. Separate clean-install command - rejected for complexity
3. Optional --clean flag - rejected for making clean behavior default

## Success Criteria
- Install command automatically cleans previous installations
- No user intervention required for state cleanup
- Clear logging of uninstall and install phases
- All existing functionality preserved
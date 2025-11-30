# Design: Uninstall with State File Implementation

## Architecture Overview

The uninstall feature will leverage the existing state file system to safely remove installed dotfiles. The design follows these principles:

1. **State-driven uninstallation**: Use the state file as the source of truth for what was installed
2. **Safety through validation**: Verify each symlink points to the expected source before removal
3. **Atomic operations**: Update state file only after successful file operations
4. **Graceful degradation**: Handle missing or corrupted state files appropriately

## Components

### 1. State File Integration
- Load existing state file from `dotfilesDir/state.yaml`
- Validate state file format and version compatibility
- Handle missing state file gracefully (inform user no tracked installations)

### 2. Symlink Validation
For each file mapping in state:
- Check if target exists and is a symlink
- Verify symlink points to the recorded source path
- Skip removal if validation fails (with warning)
- Handle cases where target was manually removed by user

### 3. File Removal Process
- Remove symlinks using `os.Remove()`
- Skip removal when symlink target doesn't match state file record
- Handle permission errors gracefully

### 4. State File Updates
- Remove successfully uninstalled entries from state file
- Save updated state file atomically
- Keep entries for files that couldn't be removed (with validation errors)

## Error Handling Strategy

1. **State file missing**: Log info message, exit gracefully
2. **State file corrupted**: Log error, suggest manual cleanup
3. **Symlink validation failures**: Log warning, skip removal, continue with others
4. **File removal failures**: Log error, continue with other files
5. **State file update failures**: Log warning but don't fail the operation

## User Experience

- Clear logging of what's being removed and why
- Summary of successful vs skipped removals
- Warnings for files that couldn't be validated or were modified
- Option for verbose output to see detailed validation results

## Security Considerations

- Never remove files that aren't recorded in state file
- Always validate symlink targets before removal
- Use absolute paths to prevent path traversal issues
- Check file permissions before attempting removal
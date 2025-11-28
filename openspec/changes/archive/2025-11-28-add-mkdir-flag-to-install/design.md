# Design for --mkdir Flag Implementation

## Architecture Overview
The `--mkdir` flag affects three layers of the system:
1. **Command Interface** (`cmd/install.go`): Flag definition and parameter passing
2. **Validation Layer** (`pkg/module/validation.go`, `pkg/module/dryrun.go`): Skip directory existence validation
3. **Installation Layer** (`pkg/module/install.go`): Create directories during symlink creation

## Design Decisions

### Flag Behavior
- **Optional flag**: `--mkdir` is optional to maintain backward compatibility
- **Mutually exclusive**: No conflicts with existing flags (`--dry-run`, `--force`)
- **Scope**: Only affects target directory creation, not source file validation

### Validation Changes
- When `--mkdir` is enabled, `ValidateTargetDirectories` should not fail on missing directories
- Directory structure validation (symlinks, permissions) should still be enforced
- File mapping validation remains unchanged

### Installation Changes
- `createSymlink` function will use `os.MkdirAll` when `--mkdir` is enabled
- Directory creation uses default permissions (0755)
- Parent directory creation is recursive as needed

### Error Handling
- If directory creation fails due to permissions, installation fails with clear error
- Symlink creation errors are unchanged
- Validation errors for other issues (conflicts, source files) remain

## Alternative Approaches Considered

### Approach 1: Always create directories
- **Pros**: Simpler implementation, no flag needed
- **Cons**: Breaking change, potential security concerns (unexpected directory creation)

### Approach 2: Configuration file option
- **Pros**: Persistent setting, per-module control
- **Cons**: More complex, requires config file changes

### Approach 3: Separate command
- **Pros**: Clear separation of concerns
- **Cons**: More commands to maintain, less convenient

**Chosen**: Flag-based approach provides user control while maintaining backward compatibility.

## Security Considerations
- Directory creation is limited to target paths specified in configuration
- No arbitrary directory creation outside configured targets
- Standard filesystem permissions apply

## Performance Impact
- Minimal impact: `os.MkdirAll` is efficient for existing directories
- Only affects installations where directories are missing
- No impact on dry-run validation performance</content>
<parameter name="filePath">/home/chao/src/dotman/openspec/changes/add-mkdir-flag-to-install/design.md
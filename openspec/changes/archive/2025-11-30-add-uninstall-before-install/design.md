# Design: Add Uninstall Before Install

## Architecture Overview

The install command will be modified to automatically execute uninstall before proceeding with installation. This creates a two-phase installation process:

1. **Cleanup Phase**: Run uninstall to remove existing managed symlinks
2. **Installation Phase**: Proceed with normal installation process

## Technical Approach

### Command Flow Integration

The uninstall call will be integrated into the install command flow at the appropriate point:

```
install command starts
    ↓
load configuration
    ↓
if not dry-run:
    run uninstall (new)
    ↓
proceed with existing install logic
```

### Key Integration Points

1. **cmd/install.go**: Add uninstall call before module.Install()
2. **pkg/module/install.go**: No changes needed (uninstall happens at higher level)
3. **pkg/module/uninstall.go**: No changes needed (reuse existing logic)

### Error Handling Strategy

- **Uninstall failures**: Log warnings but continue with installation
- **Missing state file**: Graceful handling (uninstall already does this)
- **Permission issues**: Continue with installation after logging errors

### Logging Strategy

Clear separation of uninstall and install phases in logs:

```
INFO: Starting installation
INFO: Running cleanup phase - removing previous installations
[uninstall logs...]
INFO: Cleanup phase completed
INFO: Starting installation phase
[install logs...]
```

### Dry-run Mode Considerations

- In dry-run mode, uninstall should NOT be executed
- Dry-run should only show what would happen during both phases
- Clear indication that cleanup would occur

### State Management

- Uninstall will clean the state file as part of its normal operation
- Install will create fresh state entries for the new installation
- This ensures state file always reflects current installation

### Performance Considerations

- Uninstall is typically fast (just symlink removal)
- No significant performance impact expected
- Clear user feedback about the two-phase process

### Backward Compatibility

- All existing flags (--force, --mkdir, --dry-run) work unchanged
- Install command behavior is enhanced, not changed
- Users can still call uninstall separately if desired

## Implementation Details

### Function Call Integration

In `cmd/install.go`, add uninstall call:

```go
// Add after config loading, before installation
if !dryRun {
    log.Info().Msg("Running cleanup phase - removing previous installations")
    uninstallResult, err := module.Uninstall(dotfilesDir)
    if err != nil {
        log.Warn().Err(err).Msg("Cleanup phase failed, proceeding with installation")
    } else {
        log.Info().Msg("Cleanup phase completed")
    }
}
```

### Configuration Validation

- Use same dotfilesDir for both uninstall and install
- Ensure config loading happens before uninstall (for consistency)
- No additional configuration needed

## Testing Strategy

### Unit Tests
- Test install calls uninstall when not in dry-run
- Test install skips uninstall in dry-run mode
- Test error handling when uninstall fails

### Integration Tests
- Test full install/uninstall cycle
- Test install with existing installations
- Test install with missing/corrupted state file

### Edge Cases
- Install with no previous installation
- Install with conflicting existing files
- Install with permission issues during uninstall
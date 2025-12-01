# Design: SHA1 Verification in Uninstall

## Overview
This change extends the uninstall process to handle generated files with SHA1 integrity checking, ensuring safe removal while preserving user modifications.

## Architecture Changes

### File Type Handling
- **Current**: Uninstall only processes `link` type files
- **Proposed**: Extend to process both `link` and `generated` type files
- **Rationale**: Generated files are managed by dotman and should be cleaned up during uninstall

### SHA1 Verification Flow
```
For each generated file in state:
  Calculate current SHA1 of target file
  Compare with stored SHA1 from state file
  If match: Remove file normally
  If differ: Create .bak backup, log warning, skip removal
```

### Backup Strategy
- **Naming**: Append `.bak` to original filename (e.g., `~/.bashrc` → `~/.bashrc.bak`)
- **Conflict Resolution**: If `.bak` exists, append timestamp or increment suffix
- **Location**: Backup created in same directory as original file

### State File Updates
- Successfully removed files (both links and generated) are removed from state
- Failed removals (SHA1 mismatch for generated) keep entries for manual review
- Backed up files remain in state to prevent re-installation conflicts

## Error Handling
- SHA1 calculation failures: Log warning, skip file (don't fail entire uninstall)
- Backup creation failures: Log error, skip file
- Permission issues: Continue with other files, report in summary

## Logging
- **Info**: Successful removals and backups
- **Warning**: SHA1 mismatches, backup operations
- **Error**: Permission failures, calculation errors

## Backward Compatibility
- Existing state files without SHA1 for generated files: Skip SHA1 check (treat as no verification needed)
- Symlink handling unchanged
- No breaking changes to CLI interface</content>
<parameter name="filePath">openspec/changes/add-sha1-verification-to-uninstall/design.md
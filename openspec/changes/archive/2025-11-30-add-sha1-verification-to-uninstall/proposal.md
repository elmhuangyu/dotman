# Add SHA1 Verification to Uninstall

## Summary
Extend the uninstall functionality to handle generated files (from templates) in addition to symlinks, and add SHA1 integrity checking before removal. When a generated file's content has been modified since installation, create a backup with `.bak` extension and log a warning instead of removing it.

## Motivation
Currently, uninstall only removes symlinks and validates them by checking the link target. Generated files (processed templates) are not handled during uninstall, leaving them orphaned. Additionally, there's no protection against accidental data loss if users modify generated files after installation.

By adding SHA1 verification, we ensure that only unmodified generated files are removed automatically, while modified ones are preserved with backups for user review.

## Impact
- **Safety**: Prevents accidental loss of user modifications to generated files
- **Completeness**: Uninstall now properly cleans up all managed files
- **User Experience**: Clear warnings and backups when files have been modified

## Implementation Approach
1. Extend uninstall to process `generated` type files from state file
2. For generated files, calculate current SHA1 and compare with stored value
3. If SHA1 matches, remove the file normally
4. If SHA1 differs, create `.bak` backup and log warning
5. Update state file management to handle generated file removals</content>
<parameter name="filePath">openspec/changes/add-sha1-verification-to-uninstall/proposal.md
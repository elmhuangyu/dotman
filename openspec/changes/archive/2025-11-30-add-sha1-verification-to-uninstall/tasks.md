# Tasks: Add SHA1 Verification to Uninstall

## Implementation Tasks

- [x] **Extend uninstall to process generated files**
    - Modify `Uninstall` function to handle both `link` and `generated` file types
    - Add processing loop for generated files after symlink processing
    - Update result structures to track generated file operations

- [x] **Implement SHA1 verification for generated files**
    - Add SHA1 calculation function call in uninstall validation
    - Compare calculated SHA1 with stored value from state file
    - Handle cases where SHA1 field is empty (backward compatibility)

- [x] **Add backup creation functionality**
    - Implement `createBackup` function to copy modified files with `.bak` extension
    - Handle naming conflicts (e.g., if `.bak` already exists)
    - Ensure backup is created atomically and with proper permissions

- [x] **Update validation logic**
    - Create `validateGeneratedFile` function similar to `validateSymlink`
    - Check file existence, type, and SHA1 integrity
    - Return appropriate validation results for different failure modes

- [x] **Update state file management**
    - Modify `updateStateFile` to handle removal of both link and generated file entries
    - Ensure backed up files remain in state to prevent re-installation conflicts
    - Maintain atomic state file updates

- [x] **Enhance logging and error handling**
    - Add specific log messages for SHA1 mismatches and backup operations
    - Update error collection to include backup failures
    - Ensure warnings are clear about file modifications

- [x] **Update result structures and summaries**
    - Extend `UninstallResult` to track backed up files separately
    - Modify summary generation to include backup counts
    - Update CLI output to reflect new operation types

## Testing Tasks

- [x] **Add unit tests for SHA1 verification**
    - Test SHA1 calculation and comparison logic
    - Test validation function with matching/non-matching files
    - Test edge cases (missing files, permission issues)

- [x] **Add unit tests for backup creation**
    - Test successful backup creation
    - Test conflict resolution for existing `.bak` files
    - Test backup failure scenarios

- [x] **Add integration tests for uninstall with generated files**
     - Test full uninstall flow with mixed link and generated files
     - Test SHA1 mismatch scenarios with backup creation
     - Test state file updates after partial uninstall

- [x] **Update existing tests**
     - Ensure existing symlink uninstall tests still pass
     - Add test coverage for backward compatibility with old state files

## Validation Tasks

- [x] **Manual testing scenarios**
    - Install templates, modify generated files, run uninstall
    - Verify backups are created and warnings logged
    - Confirm state file updates correctly

- [x] **Run full test suite**
    - Execute all unit and integration tests
    - Verify no regressions in existing functionality</content>
<parameter name="filePath">openspec/changes/add-sha1-verification-to-uninstall/tasks.md
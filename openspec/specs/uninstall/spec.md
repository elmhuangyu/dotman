# uninstall Specification

## Requirements

### Requirement: State File-Driven Uninstallation
The system SHALL use the state file to determine which files to uninstall and validate each file before removal, including both symlinks and generated files.

#### Scenario: Generated file uninstallation with SHA1 verification
- **WHEN** user runs `uninstall` command and state file contains generated file mappings
- **THEN** the system SHALL process generated files in addition to symlinks
- **AND** calculate SHA1 hash of current file content
- **AND** compare with stored SHA1 from state file
- **AND** remove file only if SHA1 matches
- **AND** create `.bak` backup and log warning if SHA1 differs
- **AND** update state file to remove successfully uninstalled entries

#### Scenario: SHA1 mismatch handling for generated files
- **WHEN** a generated file's current SHA1 does not match the stored SHA1
- **THEN** the system SHALL create a backup file with `.bak` extension
- **AND** log a warning message indicating the file was modified
- **AND** skip removal of the original file
- **AND** keep the entry in state file for manual review

### Requirement: File Validation Before Removal
The system SHALL validate files before removal using appropriate checks for each file type.

#### Scenario: Generated file validation
- **WHEN** processing a generated file for removal
- **THEN** the system SHALL verify the file exists and is a regular file
- **AND** calculate and verify SHA1 hash matches state file
- **AND** proceed with removal or backup based on verification result

#### Scenario: SHA1 calculation failure
- **WHEN** SHA1 calculation fails for a generated file
- **THEN** the system SHALL log a warning
- **AND** skip removal of that file
- **AND** continue processing other files

### Requirement: Backup File Creation
The system SHALL safely create backup files when generated files have been modified.

#### Scenario: Successful backup creation
- **WHEN** creating a `.bak` backup for a modified generated file
- **THEN** the system SHALL copy the file to `<original>.bak`
- **AND** handle naming conflicts by appending timestamp if needed
- **AND** ensure backup is created in the same directory as original

#### Scenario: Backup creation failure
- **WHEN** backup creation fails due to permissions or disk space
- **THEN** the system SHALL log an error
- **AND** skip removal but keep the original file
- **AND** include the failure in the operation summary

### Requirement: Error Handling and Logging
The system SHALL provide clear feedback about uninstallation operations including backup operations.

#### Scenario: Comprehensive operation summary
- **WHEN** uninstallation completes
- **THEN** the system SHALL display a summary showing:
  - Number of symlinks successfully removed
  - Number of generated files successfully removed
  - Number of files backed up due to modifications
  - Number of files skipped due to validation failures
  - Number of files that were already missing
  - Any errors encountered during the process
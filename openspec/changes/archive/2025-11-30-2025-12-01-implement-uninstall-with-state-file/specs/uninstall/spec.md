# uninstall Specification

## Purpose
Define the behavior and requirements for uninstalling dotfiles using the state file to ensure safe and tracked removal of installed files.

## ADDED Requirements

### Requirement: State File-Driven Uninstallation
The system SHALL use the state file to determine which files to uninstall and validate each symlink before removal.

#### Scenario: Successful uninstallation with valid state file
- **WHEN** user runs `uninstall` command and a valid state file exists
- **THEN** the system SHALL load the state file from `dotfilesDir/state.yaml`
- **AND** iterate through all file mappings in the state file
- **AND** validate each symlink points to the recorded source before removal
- **AND** remove only symlinks that pass validation
- **AND** update the state file to remove successfully uninstalled entries
- **AND** return success with summary of removed files

#### Scenario: State file missing
- **WHEN** user runs `uninstall` command and no state file exists
- **THEN** the system SHALL log an informational message
- **AND** exit gracefully without error
- **AND** inform user that no tracked installations were found

#### Scenario: State file corrupted or invalid format
- **WHEN** user runs `uninstall` command and state file is corrupted
- **THEN** the system SHALL log an error message
- **AND** suggest manual cleanup or state file repair
- **AND** exit with error status

### Requirement: Symlink Validation Before Removal
The system SHALL validate that each target file is a symlink pointing to the expected source before removal.

#### Scenario: Valid symlink validation
- **WHEN** a target file exists and is a symlink pointing to the recorded source
- **THEN** the system SHALL proceed with removal
- **AND** log the removal operation

#### Scenario: Invalid symlink target
- **WHEN** a target file exists but is a symlink pointing to a different source
- **THEN** the system SHALL skip removal with a warning
- **AND** log that the symlink appears to have been modified manually
- **AND** keep the entry in the state file for manual review

#### Scenario: Target is not a symlink
- **WHEN** a target file exists but is not a symlink (regular file or directory)
- **THEN** the system SHALL skip removal with a warning
- **AND** log that the target is not a managed symlink
- **AND** keep the entry in the state file

#### Scenario: Target file missing
- **WHEN** a target file from the state file does not exist
- **THEN** the system SHALL skip removal with an info message
- **AND** remove the entry from the state file (since it's already gone)

### Requirement: Skip Modified Symlinks
The system SHALL skip removal of symlinks that don't match the state file record and log warnings.

#### Scenario: Symlink points to different source
- **WHEN** a target symlink exists but points to a different source than recorded in state
- **THEN** the system SHALL skip removal with a warning log
- **AND** inform user that the symlink appears to have been modified
- **AND** keep the entry in the state file for manual review

### Requirement: State File Management
The system SHALL properly update the state file after uninstallation operations.

#### Scenario: Successful state file update
- **WHEN** uninstallation operations complete successfully
- **THEN** the system SHALL remove entries for successfully uninstalled files
- **AND** keep entries for files that couldn't be removed due to validation failures
- **AND** save the updated state file atomically

#### Scenario: State file update failure
- **WHEN** saving the updated state file fails
- **THEN** the system SHALL log a warning
- **AND** not fail the uninstallation operation
- **AND** inform user that state file couldn't be updated

### Requirement: Error Handling and Logging
The system SHALL provide clear feedback about uninstallation operations and handle errors gracefully.

#### Scenario: Permission denied during removal
- **WHEN** attempting to remove a symlink fails due to permissions
- **THEN** the system SHALL log an error
- **AND** continue with other files
- **AND** include the failure in the final summary

#### Scenario: Comprehensive operation summary
- **WHEN** uninstallation completes
- **THEN** the system SHALL display a summary showing:
  - Number of files successfully removed
  - Number of files skipped due to validation failures
  - Number of files that were already missing
  - Any errors encountered during the process
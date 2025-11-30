# uninstall Specification

## Purpose
Define the behavior and requirements for uninstalling dotfiles when called as part of the install command cleanup phase.

## MODIFIED Requirements

### Requirement: Uninstall as Install Precondition [MODIFIED]
The uninstall command SHALL be callable by the install command as a cleanup phase before new installation.

#### Scenario: Uninstall called during install cleanup
- **WHEN** install command calls uninstall as cleanup phase
- **THEN** the system SHALL execute standard uninstall behavior
- **AND** use the same dotfiles directory as the subsequent installation
- **AND** handle missing state file gracefully (no previous installation)
- **AND** provide appropriate logging for the cleanup context

#### Scenario: Uninstall error handling in install context
- **WHEN** uninstall encounters errors during install cleanup phase
- **THEN** the system SHALL log errors and warnings as usual
- **AND** not prevent the install command from proceeding
- **AND** return error information to the install caller for logging

#### Scenario: Uninstall logging in install context
- **WHEN** uninstall runs as part of install cleanup
- **THEN** the system SHALL provide standard uninstall logging
- **AND** integrate with install command's phase separation
- **AND** clearly indicate which operations are part of cleanup

## Existing Requirements (Unchanged)

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

### Requirement: Symlink Validation Before Removal
The system SHALL validate that each target file is a symlink pointing to the expected source before removal.

### Requirement: Skip Modified Symlinks
The system SHALL skip removal of symlinks that don't match the state file record and log warnings.

### Requirement: State File Management
The system SHALL properly update the state file after uninstallation operations.

### Requirement: Error Handling and Logging
The system SHALL provide clear feedback about uninstallation operations and handle errors gracefully.
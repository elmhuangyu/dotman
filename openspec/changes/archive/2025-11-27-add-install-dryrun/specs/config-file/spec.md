## ADDED Requirements
### Requirement: File Conflict Validation
The system SHALL validate that no conflicts exist in the file mapping during dry-run mode.

#### Scenario: Target file exists with correct symlink
- **WHEN** a target file exists and points to the expected source file
- **THEN** the system SHALL consider this valid
- **AND** report no conflict

#### Scenario: Target file exists with wrong symlink
- **WHEN** a target file exists but points to a different source file
- **THEN** the system SHALL detect the conflict
- **AND** report the target file exists with wrong symlink
- **AND** indicate what it currently points to

#### Scenario: Target file exists as regular file
- **WHEN** a target file exists as a regular file (not a symlink)
- **THEN** the system SHALL detect the conflict
- **AND** report the target file exists and would be overwritten

### Requirement: Directory Structure Validation
The system SHALL ensure all target parent directories are valid (directories, not symlinks).

#### Scenario: Parent directory validation
- **WHEN** validating target file paths
- **THEN** the system SHALL check all parent directories exist as directories
- **AND** ensure no parent path component is a symlink
- **AND** report validation errors for invalid directory structure

#### Scenario: Target directory validation
- **WHEN** validating target directories from Dotfile configuration
- **THEN** the system SHALL ensure target directories are directories not symlinks
- **AND** report validation errors for invalid target directories
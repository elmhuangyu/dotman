# install Specification

## Purpose
Define the behavior and requirements for installing dotfiles with automatic cleanup of previous installations.

## MODIFIED Requirements

### Requirement: Install-Uninstall Integration [MODIFIED]
The install command SHALL automatically run uninstall before installation to ensure a clean state.

#### Scenario: Automatic cleanup before installation
- **WHEN** user runs `install` command without --dry-run flag
- **THEN** the system SHALL first run uninstall to remove existing managed symlinks
- **AND** log the cleanup phase separately from installation phase
- **AND** proceed with normal installation after cleanup completes
- **AND** continue with installation even if cleanup encounters errors

#### Scenario: Dry-run mode skips cleanup
- **WHEN** user runs `install --dry-run` command
- **THEN** the system SHALL NOT run uninstall
- **AND** only show what would be installed without cleanup
- **AND** indicate that cleanup would occur in normal mode

#### Scenario: Cleanup phase error handling
- **WHEN** uninstall encounters errors during cleanup phase
- **THEN** the system SHALL log warnings for cleanup failures
- **AND** continue with installation process
- **AND** not fail the entire install command due to cleanup issues

#### Scenario: Cleanup phase logging
- **WHEN** install command runs cleanup phase
- **THEN** the system SHALL log "Running cleanup phase - removing previous installations"
- **AND** log "Cleanup phase completed" when uninstall finishes
- **AND** log "Starting installation phase" before beginning installation
- **AND** provide clear separation between cleanup and installation phases

#### Scenario: No previous installation cleanup
- **WHEN** no previous installation exists (no state file)
- **THEN** the system SHALL run uninstall which handles missing state gracefully
- **AND** continue with installation without errors
- **AND** log appropriate messages from uninstall phase

#### Scenario: Cleanup with existing flags
- **WHEN** user runs `install --force` or `install --mkdir` with cleanup
- **THEN** the system SHALL run uninstall first with standard behavior
- **AND** apply force/mkdir flags during installation phase only
- **AND** maintain existing flag behaviors for installation

## Existing Requirements (Unchanged)

### Requirement: Install Execution
The system SHALL create symlinks from source dotfiles to target locations when running the install command without --dry-run.

### Requirement: --mkdir Flag Support
The install command SHALL support a `--mkdir` flag to automatically create missing target directories.

### Requirement: --force Flag Support
The install command SHALL support a `--force` flag to handle conflicting target files by backing them up and proceeding with installation.

### Requirement: Install State Logging
The system SHALL record all successful symlink installations in a state file for tracking and management purposes.
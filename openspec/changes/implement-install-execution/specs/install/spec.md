## ADDED Requirements
### Requirement: Install Execution
The system SHALL create symlinks from source dotfiles to target locations when running the install command without --dry-run.

#### Scenario: Successful installation
- **WHEN** user runs `install` with valid configuration and no conflicts
- **THEN** the system SHALL create symlinks for all file mappings
- **AND** skip creation when correct symlinks already exist
- **AND** return success when all symlinks are created

#### Scenario: Installation with conflicts
- **WHEN** user runs `install` and conflicts are detected
- **THEN** the system SHALL return an error
- **AND** not create any symlinks
- **AND** report the specific conflicts that prevent installation

#### Scenario: Skip correct symlinks
- **WHEN** a target location already has the correct symlink
- **THEN** the system SHALL skip creating the symlink
- **AND** log that the symlink already exists correctly

#### Scenario: Installation failure handling
- **WHEN** symlink creation fails for any reason
- **THEN** the system SHALL return an error
- **AND** stop the installation process
- **AND** report the specific failure reason
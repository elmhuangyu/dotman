# install Specification

## Purpose
Define the behavior and requirements for installing dotfiles by creating symlinks from source files to target locations.

## Requirements
### Requirement: Install Execution
The system SHALL create symlinks from source dotfiles to target locations when running the install command without --dry-run.

#### Scenario: Installation with missing directories and --mkdir flag
- **WHEN** user runs `install --mkdir` and target directories do not exist
- **THEN** the system SHALL create missing target directories automatically
- **AND** create symlinks for all file mappings
- **AND** return success when all directories and symlinks are created

#### Scenario: Installation with missing directories without --mkdir flag
- **WHEN** user runs `install` without --mkdir and target directories do not exist
- **THEN** the system SHALL return an error
- **AND** not create any symlinks
- **AND** report the missing directories

### Requirement: --mkdir Flag Support
The install command SHALL support a `--mkdir` flag to automatically create missing target directories.

#### Scenario: Flag availability
- **WHEN** user runs `install --help`
- **THEN** the system SHALL show the `--mkdir` flag in the help output
- **AND** describe its purpose as creating missing directories

#### Scenario: Flag validation
- **WHEN** user specifies both `--dry-run` and `--mkdir`
- **THEN** the system SHALL accept both flags
- **AND** perform dry-run validation without directory checks when `--mkdir` is specified

#### Scenario: Directory creation permissions
- **WHEN** creating directories with `--mkdir`
- **THEN** the system SHALL use default directory permissions (0755)
- **AND** create parent directories recursively as needed

#### Scenario: Directory creation failure
- **WHEN** directory creation fails due to permissions or other filesystem errors
- **THEN** the system SHALL return an error
- **AND** stop the installation process
- **AND** report the specific directory creation failure

### Requirement: --force Flag Support
The install command SHALL support a `--force` flag to handle conflicting target files by backing them up and proceeding with installation.

#### Scenario: Force flag availability
- **WHEN** user runs `install --help`
- **THEN** the system SHALL show the `--force` flag in the help output
- **AND** describe its purpose as forcing installation by overwriting existing files

#### Scenario: Force flag validation
- **WHEN** user specifies both `--dry-run` and `--force` flags
- **THEN** the system SHALL reject the combination with an error
- **AND** display a message indicating only one of --dry-run or --force can be used

#### Scenario: Module config conflicts in force mode
- **WHEN** multiple source files map to the same target (module config conflict)
- **AND** user runs `install --force`
- **THEN** the system SHALL still fail with an error
- **AND** not perform any installation

#### Scenario: Target file conflicts in force mode
- **WHEN** target files exist as regular files or wrong symlinks
- **AND** user runs `install --force`
- **THEN** the system SHALL backup existing files with .bak extension
- **AND** create correct symlinks to replace them
- **AND** return success when all operations complete

#### Scenario: Backup file naming
- **WHEN** backing up conflicting files in force mode
- **THEN** the system SHALL append .bak extension to the original filename
- **AND** overwrite any existing .bak file if present

#### Scenario: Force mode with mkdir
- **WHEN** user runs `install --force --mkdir`
- **THEN** the system SHALL enable both force and mkdir behaviors
- **AND** create missing directories as needed
- **AND** backup conflicting files before creating symlinks

### Requirement: Install State Logging
The system SHALL record all successful symlink installations in a state file for tracking and management purposes.

#### Scenario: State file creation on first install
- **WHEN** user runs install command and no state file exists
- **THEN** the system SHALL create a new state file with version information
- **AND** record all successfully created symlinks with source, target, and type information

#### Scenario: State file update on subsequent installs
- **WHEN** user runs install command and state file exists
- **THEN** the system SHALL load existing state file
- **AND** append newly created symlinks to the existing file mappings
- **AND** save the updated state file atomically

#### Scenario: State file recording for each symlink
- **WHEN** a symlink is successfully created during installation
- **THEN** the system SHALL immediately record the file mapping in state
- **AND** include source path, target path, and type "link"
- **AND** ensure the state entry persists even if later symlinks fail

#### Scenario: State file location and format
- **WHEN** the install command records state
- **THEN** the system SHALL store the state file in a predictable location
- **AND** use YAML format matching the StateFile struct
- **AND** include version compatibility information

#### Scenario: Error handling for state file operations
- **WHEN** state file operations fail during installation
- **THEN** the system SHALL log a warning but continue with installation
- **AND** not fail the entire installation process due to state logging issues

## MODIFIED Requirements

### Requirement: State File Cleanup Integration
The system SHALL maintain state file entries that can be used by the uninstall command for safe removal.

#### Scenario: State file persistence for uninstall
- **WHEN** install command records symlinks in state file
- **THEN** the system SHALL ensure entries contain complete source and target paths
- **AND** use absolute paths to enable reliable validation during uninstall
- **AND** maintain file type information to distinguish between links and generated files

#### Scenario: State file consistency after failed installs
- **WHEN** installation partially fails after some symlinks are created
- **THEN** the system SHALL preserve state entries for successfully created symlinks
- **AND** enable uninstall command to safely remove only the files that were actually installed
- **AND** maintain state file integrity for subsequent operations
# install Specification Delta

## MODIFIED Requirements
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

## ADDED Requirements
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
- **AND** report the specific directory creation failure</content>
<parameter name="filePath">/home/chao/src/dotman/openspec/changes/add-mkdir-flag-to-install/specs/install/spec.md
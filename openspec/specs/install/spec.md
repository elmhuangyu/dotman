# install Specification

## Purpose
TBD - created by archiving change implement-install-execution. Update Purpose after archive.
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
- **AND** report the specific directory creation failure</content>
<parameter name="filePath">/home/chao/src/dotman/openspec/changes/add-mkdir-flag-to-install/specs/install/spec.md

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
- **AND** backup conflicting files before creating symlinks</content>
<parameter name="filePath">/home/chao/src/dotman/openspec/changes/implement-force-flag-for-install/specs/install/spec.md

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

### Requirement: Install-Uninstall Integration [MODIFIED]
The install command SHALL create state file entries that enable safe and tracked uninstallation of dotfiles.

#### Scenario: State file entries for uninstall tracking
- **WHEN** install command creates symlinks
- **THEN** system SHALL record absolute paths for both source and target in state file
- **AND** include file type as "link" to distinguish from generated files
- **AND** ensure entries are sufficient for uninstall validation

#### Scenario: State file persistence for uninstall
- **WHEN** install command completes (both successful and with skipped symlinks)
- **THEN** system SHALL ensure state file contains all tracked symlinks
- **AND** maintain state file integrity for subsequent uninstall operations
- **AND** handle state file corruption gracefully during uninstall

#### Scenario: Uninstall command dependency
- **WHEN** uninstall command is executed
- **THEN** system SHALL rely on state file entries created by install command
- **AND** validate symlinks against recorded source paths before removal
- **AND** only remove symlinks that match state file records exactly

### Requirement: Template File Processing
The system SHALL process files with `.dot-tmpl` extension as Go text templates and generate rendered files instead of creating symlinks.

#### Scenario: Template file detection and processing
- **WHEN** install command encounters a file with `.dot-tmpl` extension in a module
- **THEN** the system SHALL process it as a Go text template
- **AND** use `root_config.Vars` as template data
- **AND** generate a rendered file at the target location (without `.dot-tmpl` extension)
- **AND** not create a symlink for template files

#### Scenario: Template rendering with variables
- **WHEN** processing a `.dot-tmpl` file
- **AND** the template contains variables like `{{.VAR_NAME}}`
- **THEN** the system SHALL replace variables with corresponding values from `root_config.Vars`
- **AND** generate the file with rendered content
- **AND** report success when template rendering completes

#### Scenario: Template file naming
- **WHEN** processing a template file named `config.dot-tmpl`
- **THEN** the system SHALL generate the target file named `config` (without `.dot-tmpl` extension)
- **AND** place it in the configured target location
- **AND** use the generated file name for state tracking

#### Scenario: Mixed template and regular files
- **WHEN** a module contains both `.dot-tmpl` files and regular files
- **THEN** the system SHALL process template files by generating rendered files
- **AND** process regular files by creating symlinks
- **AND** handle both file types in the same install operation

#### Scenario: Template syntax errors
- **WHEN** a `.dot-tmpl` file contains invalid Go template syntax
- **THEN** the system SHALL fail the installation
- **AND** report the specific template syntax error
- **AND** not create any files for that template

#### Scenario: Missing template variables
- **WHEN** a template references a variable not present in `root_config.Vars`
- **THEN** the system SHALL fail the installation
- **AND** report the missing variable name
- **AND** not create the generated file

#### Scenario: Template file with --dry-run
- **WHEN** user runs install with --dry-run flag
- **AND** there are `.dot-tmpl` files in modules
- **THEN** the system SHALL validate template syntax
- **AND** verify all required variables are available
- **AND** report what files would be generated without actually creating them

#### Scenario: Template file with --mkdir flag
- **WHEN** processing `.dot-tmpl` files with --mkdir flag
- **AND** target directories do not exist
- **THEN** the system SHALL create missing target directories
- **AND** generate the rendered template files in the created directories

#### Scenario: Template file with --force flag
- **WHEN** processing `.dot-tmpl` files with --force flag
- **AND** target files already exist
- **THEN** the system SHALL backup existing files with .bak extension
- **AND** generate new template-rendered files to replace them


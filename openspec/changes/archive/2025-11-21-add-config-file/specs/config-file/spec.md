## ADDED Requirements
### Requirement: Module Configuration File
The system SHALL support a "Dotfile" YAML configuration file in module directories to specify target destinations for files.

#### Scenario: Config file discovery
- **WHEN** processing a module directory
- **THEN** the system SHALL check for a "Dotfile" YAML file in that directory
- **AND** use the configuration if present

#### Scenario: Target directory specification
- **WHEN** a "Dotfile" contains a `target_dir` field
- **THEN** the system SHALL map all files from the module directory to the specified target directory
- **AND** create symbolic links from source files to target locations

### Requirement: Configuration File Format
The "Dotfile" configuration SHALL use YAML format with a `target_dir` field.

#### Scenario: Valid config file format
- **WHEN** a "Dotfile" contains valid YAML with `target_dir` field
- **THEN** the system SHALL successfully parse and use the configuration
- **EXAMPLE**:
```yaml
target_dir: "/home/user/.config/nvim"
```

#### Scenario: Invalid config file handling
- **WHEN** a "Dotfile" contains invalid YAML or missing required fields
- **THEN** the system SHALL provide clear error messages
- **AND** continue processing other modules if possible

### Requirement: Flag Precedence
The system SHALL prioritize command-line flags over configuration file settings when both are present.

#### Scenario: Flag overrides config
- **WHEN** both `--dir` flag and config file `target_dir` are specified
- **THEN** the system SHALL use the `--dir` flag value
- **AND** log that flag is overriding config file setting
## ADDED Requirements
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
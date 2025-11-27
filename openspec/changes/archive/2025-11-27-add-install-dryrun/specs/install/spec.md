## ADDED Requirements
### Requirement: Install Dry Run Mode
The system SHALL provide a --dry-run flag for the install command that validates installation without making filesystem changes.

#### Scenario: Dry run validation success
- **WHEN** user runs `install --dry-run` with valid configuration
- **THEN** the system SHALL build file mappings and validate all installation constraints
- **AND** display what would be installed without making changes
- **AND** return success if all validations pass

#### Scenario: Dry run conflict detection
- **WHEN** two source files would map to the same target location
- **THEN** the system SHALL detect the conflict
- **AND** report an error indicating the duplicate target
- **AND** not make any filesystem changes

## MODIFIED Requirements
### Requirement: Install Command Flags
The system SHALL support --dry-run and --force flags for the install command with mutually exclusive validation.

#### Scenario: Flag validation
- **WHEN** both --dry-run and --force flags are provided
- **THEN** the system SHALL return an error indicating flags are mutually exclusive
- **AND** display usage information
- **AND** not proceed with installation
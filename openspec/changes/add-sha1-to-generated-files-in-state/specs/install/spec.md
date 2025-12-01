# install Specification Delta

## MODIFIED Requirements

### Requirement: Template File Processing
The system SHALL process files with `.dot-tmpl` extension as Go text templates and generate rendered files instead of creating symlinks.

#### ADDED Scenario: SHA1 calculation for generated files
- **WHEN** a template file is successfully processed and generated
- **THEN** the system SHALL calculate SHA1 hash of the generated file content
- **AND** record the SHA1 hash in the state file mapping for integrity verification
- **AND** store the hash as a hex-encoded string in the `sha1` field

#### ADDED Scenario: State file SHA1 field population
- **WHEN** recording generated file mappings in state file
- **THEN** the system SHALL populate the `sha1` field with calculated hash
- **AND** leave `sha1` field empty for link type mappings
- **AND** ensure SHA1 calculation does not fail the installation process</content>
<parameter name="filePath">openspec/changes/add-sha1-to-generated-files-in-state/specs/install/spec.md
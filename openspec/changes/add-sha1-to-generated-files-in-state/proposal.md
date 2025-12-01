# Add SHA1 Calculation for Generated Files in State

## Summary
Modify the state file handling to calculate and store SHA1 hashes for files of type "generated" (template-generated files). This enables integrity checking and change detection for generated files.

## Motivation
Currently, the state file records generated files with type "generated" but does not store any integrity information. Adding SHA1 calculation allows:
- Verification that generated files haven't been modified externally
- Detection of changes in source templates or variables that would affect the generated content
- Better tracking of file state for uninstall and update operations

## Impact
- State files will include SHA1 field for generated file mappings
- Install operations will compute SHA1 after generating template files
- Backward compatibility maintained (SHA1 is optional field)

## Implementation Approach
- Modify `AddFileMapping` in `pkg/state/state_file.go` to calculate SHA1 when type is "generated"
- SHA1 computed from the target file content after generation
- Update install logic to ensure SHA1 is calculated after file creation</content>
<parameter name="filePath">openspec/changes/add-sha1-to-generated-files-in-state/proposal.md
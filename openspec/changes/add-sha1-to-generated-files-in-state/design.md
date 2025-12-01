# Design for SHA1 Calculation in State Files

## Architecture Decision
The SHA1 calculation will be performed immediately after file generation in the install process. This ensures the hash reflects the exact content written to disk.

## Trade-offs
- **Performance**: SHA1 calculation adds I/O overhead for reading generated files. However, this is acceptable since template generation is typically infrequent and file sizes are small.
- **Error Handling**: If SHA1 calculation fails, we log a warning but continue installation. This prevents SHA1 failures from breaking installs.
- **Backward Compatibility**: SHA1 field is optional in YAML, so existing state files remain valid.

## Implementation Details
- SHA1 computed using Go's crypto/sha1 package
- Hex encoding for storage in YAML
- Calculation only for `TypeGenerated` files, not for symlinks (`TypeLink`)
- State file format remains compatible (SHA1 field was already defined as optional)</content>
<parameter name="filePath">openspec/changes/add-sha1-to-generated-files-in-state/design.md
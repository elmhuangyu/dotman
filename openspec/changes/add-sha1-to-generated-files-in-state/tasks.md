# Tasks for Adding SHA1 to Generated Files in State

- [x] **Add SHA1 calculation function to state package**
   - Implement `calculateSHA1(filePath string) (string, error)` in `pkg/state/state_file.go`
   - Use crypto/sha1 to compute hash of file content
   - Return hex-encoded string

- [x] **Modify AddFileMapping to calculate SHA1 for generated files**
   - Update `AddFileMapping` method to accept target file path for SHA1 calculation
   - When `fileType == TypeGenerated`, compute SHA1 of target file and set in mapping
   - Handle errors gracefully (log warning but continue)

- [x] **Update install.go to pass target path for SHA1 calculation**
   - Modify calls to `AddFileMapping` for generated files to ensure target file exists before calling
   - Verify SHA1 is computed after template file creation

- [x] **Add unit tests for SHA1 calculation**
   - Test SHA1 calculation function with known file content
   - Test AddFileMapping sets SHA1 correctly for generated type
   - Test error handling when target file doesn't exist

- [x] **Update state file tests**
   - Modify existing tests to verify SHA1 field is populated for generated mappings
   - Ensure backward compatibility (existing state files without SHA1 still load)

- [x] **Validate with openspec**
   - Run `openspec validate add-sha1-to-generated-files-in-state --strict`
   - Resolve any validation issues</content>
<parameter name="filePath">openspec/changes/add-sha1-to-generated-files-in-state/tasks.md
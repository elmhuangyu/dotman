# Implementation Tasks for --mkdir Flag

## 1. Command Interface Changes
- [ ] Add `mkdirFlag bool` variable to `cmd/install.go`
- [ ] Add `--mkdir` flag definition in `init()` function
- [ ] Update `PreRunE` validation to handle `--mkdir` flag combinations
- [ ] Pass `mkdirFlag` to `install()` function
- [ ] Update function signature: `install(dotfilesDir string, dryRun, force, mkdir bool)`

## 2. Validation Layer Updates
- [ ] Modify `Validate()` in `pkg/module/dryrun.go` to accept `mkdir` parameter
- [ ] Update `ValidateTargetDirectories()` to skip missing directory errors when `mkdir=true`
- [ ] Ensure other validation logic (symlinks, permissions) still runs
- [ ] Update dry-run mode to reflect directory creation when `--mkdir` is used

## 3. Installation Layer Updates
- [ ] Modify `Install()` in `pkg/module/install.go` to accept `mkdir` parameter
- [ ] Update `createSymlink()` to accept `mkdir` parameter
- [ ] Replace directory existence check with `os.MkdirAll()` when `mkdir=true`
- [ ] Maintain existing error handling for other symlink creation failures

## 4. Testing Implementation
- [ ] Add unit tests for `createSymlink` with `mkdir=true`
- [ ] Add integration tests for `install --mkdir` command
- [ ] Test directory creation with nested missing directories
- [ ] Test error handling when directory creation fails due to permissions
- [ ] Ensure existing tests still pass without `--mkdir` flag

## 5. Documentation and Validation
- [ ] Update command help text for `--mkdir` flag
- [ ] Run `openspec validate add-mkdir-flag-to-install --strict` to ensure spec compliance
- [ ] Test manual installation scenarios to verify functionality</content>
<parameter name="filePath">/home/chao/src/dotman/openspec/changes/add-mkdir-flag-to-install/tasks.md
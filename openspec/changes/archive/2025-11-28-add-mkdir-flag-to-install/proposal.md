# Add --mkdir Flag to Install Command

## Problem
Currently, the `install` command fails if any target directories do not exist. Users must manually create all required directories before running the install command, which is inconvenient for setting up dotfiles on new systems.

## Solution
Add a `--mkdir` flag to the `install` subcommand that automatically creates missing target directories during installation, eliminating the need for manual directory creation.

## Impact
- **Scope**: Affects the install command behavior and validation logic
- **Breaking Changes**: None - this is an additive feature with a new optional flag
- **Risk**: Low - the flag is optional and only affects directory creation behavior

## Implementation Approach
1. Add `--mkdir` flag to the install command in `cmd/install.go`
2. Modify validation logic to skip directory existence checks when `--mkdir` is enabled
3. Update `createSymlink` function to create missing directories using `os.MkdirAll` when `--mkdir` is enabled
4. Add tests for the new functionality

## Dependencies
None - this is a self-contained feature addition.

## Testing
- Unit tests for directory creation logic
- Integration tests with `--mkdir` flag
- Ensure existing behavior is preserved when flag is not used</content>
<parameter name="filePath">/home/chao/src/dotman/openspec/changes/add-mkdir-flag-to-install/proposal.md
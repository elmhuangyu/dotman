# Change: Implement Install Execution Logic

## Why
The install command currently only performs dry-run validation but doesn't actually install the dotfiles. Users need the ability to create symlinks from their dotfiles to the target locations after validation passes.

## What Changes
- Implement actual symlink creation logic in pkg/module package
- Add Install function that creates symlinks for validated mappings
- Skip creation when correct symlinks already exist
- Return errors for any conflicts or installation failures
- Update install command to call the new installation logic
- Ignore force flag for now (conflicts will cause errors)

## Impact
- Affected specs: install
- Affected code: cmd/install.go, pkg/module/ (new install.go)
- Enables actual dotfile installation functionality</content>
<parameter name="filePath">/home/chao/src/dotman/openspec/changes/implement-install-execution/proposal.md
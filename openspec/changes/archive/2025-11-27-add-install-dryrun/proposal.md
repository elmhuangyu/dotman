# Change: Add Install Dry Run Feature

## Why
The install command needs a --dry-run flag to show users what actions would be performed without actually making changes to their filesystem, providing safety and visibility before installation.

## What Changes
- Add --dry-run validation logic to pkg/module package
- Implement file conflict detection and validation
- Build two-way mapping of source to target files
- Ensure no two files map to the same target
- Validate target paths don't exist or have correct symlinks
- Check target directories and parent paths are directories (not symlinks)
- Extend install command to support dry-run mode

## Impact
- Affected specs: install, config-file
- Affected code: cmd/install.go, pkg/module/ (new validation package)
- New safety features for the installation process
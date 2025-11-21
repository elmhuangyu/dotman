# Change: Add Module Config File Support

## Why
Users need to specify where files in each dotfile module should be mapped to in the target system. Currently there's no way to define target destinations for files within a module directory.

## What Changes
- Add support for a "Dotfile" YAML configuration file in module directories
- Add `target_dir` field to specify where files in the current module directory should be mapped to
- Config file will be discovered in the current module directory being processed
- This enables per-module configuration for file mapping destinations
- **BREAKING**: None - this is additive functionality

## Impact
- Affected specs: None (new capability)
- Affected code:
  - `cmd/install.go` (install command implementation)
  - `cmd/uninstall.go` (uninstall command implementation)
  - `cmd/verify.go` (verify command implementation)
  - Add new config parsing logic for reading Dotfile YAML
  - Module discovery and processing logic
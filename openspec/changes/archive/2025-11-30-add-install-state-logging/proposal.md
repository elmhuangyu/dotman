# Change: Add Install State Logging

## Why
To track and persist the state of installed dotfiles for better management, rollback capabilities, and visibility into what files have been installed by the system.

## What Changes
- Add state file logging when install command creates symlinks
- Record successful installations in the state file with source, target, and type information
- Update state file after each successful symlink creation
- Handle state file creation and updates atomically

## Impact
- Affected specs: install
- Affected code: pkg/module/install.go, pkg/state/state_file.go
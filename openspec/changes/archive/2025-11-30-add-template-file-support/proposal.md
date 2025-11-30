# Change: Add Template File Support for Install Command

## Why
Enable users to generate configuration files from templates using environment variables, providing more flexibility than static symlinks for dynamic configurations.

## What Changes
- Add support for `.dot-tmpl` files in modules to be processed as Go text templates
- Use `root_config.Vars` as template data for rendering
- Generate rendered files instead of symlinks for template files
- Maintain existing symlink behavior for all other file types

## Impact
- Affected specs: install
- Affected code: pkg/module/install.go, pkg/config/root_config.go (template data), potentially pkg/module/file_mapping.go
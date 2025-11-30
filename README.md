# Dot Installer

A simple command-line tool to manage your dotfiles.

## Description

Dot Installer helps you manage your dotfiles by installing and uninstalling them based on a configuration file. It creates symbolic links from your dotfiles directory to your home directory, making it easy to keep your dotfiles in a version-controlled repository.

## Why not GNU Stow?

While GNU Stow is a popular tool for managing dotfiles, it has several limitations that motivated the creation of this project:

| Stow Limitation | Our Solution |
|---|---|
| **Directory symlinks** - Stow creates directory-level symlinks, which can lead to software modifications polluting your dotfile repository | **File-only linking** - Only creates file-level symbolic links, preventing accidental pollution of your dotfiles |
| **No environment-specific configs** - Cannot handle different configurations for different environments (Linux/macOS/work/home) | **Template system** - Supports environment-specific configurations through templating and conditional deployment |
| **Unreliable uninstallation** - `unstow` cannot accurately track which files were originally installed | **State tracking** - Maintains a manifest of installed files for reliable uninstallation |

## Installation

To install the tool, you can clone this repository and build it using Go:

```bash
git clone https://github.com/elmhuangyu/dotman.git
cd dotman
go build -o dotman ./cmd/main
```

## Usage

The CLI tool provides professional subcommand handling with Cobra and structured logging with zerolog.

### Global Flags

- `--debug`: Enable debug logging for verbose output
- `--dir <path>`: Specify custom dotfiles directory (default: `$HOME/.config/dotfiles`)

### Configuration

dotman supports modular dotfile management using "Dotfile" configuration files. Each module directory can contain a `Dotfile` YAML that specifies where files should be mapped.

#### Example Directory Structure

```
~/.config/dotfiles/
├── DotRoot                 # Root configuration file
├── nvim/
│   ├── init.vim
│   ├── Dotfile
│   └── lua/
│       └── config.lua
├── git/
│   ├── .gitconfig
│   └── Dotfile
├── bash/
│   ├── .bashrc
│   └── Dotfile
├── temp/                   # Will be excluded by DotRoot, see below
│   └── Dotfile
└── backup/                 # Will be excluded by DotRoot, see below
    └── old-config/
```

#### Root Configuration (DotRoot)

You can create a `DotRoot` file in your dotfiles root directory to configure global settings:

```yaml
# ~/.config/dotfiles/DotRoot
vars:
  USERNAME: "john"
  HOMEDIR: "/home/john"
  VERSION: "1.0.0"
exclude_modules:
  - "temp"
  - "backup"
  - "old-config"
```

**Root Configuration Fields:**
- `vars`: Define variables that can be used in module configurations (currently for future templating support)
- `exclude_modules`: List of module directory names to skip during installation


#### Dotfile Configuration Format

Each module can contain a `Dotfile` YAML configuration:

```yaml
# nvim/Dotfile
target_dir: "/home/user/.config/nvim"
```

### Commands

#### `install`

The `install` subcommand creates symbolic links for your dotfiles based on module configurations.
It automatically runs a cleanup phase (uninstall) before installation to ensure a clean state
and prevent conflicts from previous installations.

```bash
# Basic usage - installs all modules with Dotfile configurations
dotman install

# With debug logging to see detailed operations
dotman --debug install

# With custom dotfiles directory
dotman --dir /path/to/dotfiles install

# Force installation (overwrite existing files)
dotman install --force

# Create missing target directories
dotman install --mkdir

# Dry-run mode (show what would be installed without making changes)
dotman install --dry-run
```

#### `uninstall`

The `uninstall` subcommand removes symbolic links created by dotman, safely leaving other files untouched.

```bash
dotman uninstall

# With debug mode
dotman --debug uninstall
```


#### Getting Help

```bash
# Show main help
dotman --help

# Show help for a specific command
dotman install --help
```

### Features

- **Modular Configuration**: Each module has its own `Dotfile` configuration
- **Safe Installation**: Only creates symbolic links, never overwrites files without warning
- **Safe Uninstallation**: Only removes links created by dotman
- **Detailed Logging**: Debug mode provides detailed operation information

## License

This project is licensed under the Apache-2 License.

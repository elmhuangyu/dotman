# Dot Installer

A simple command-line tool to manage your dotfiles.

## Description

Dot Installer helps you manage your dotfiles by installing, uninstalling, and verifying them based on a configuration file. It creates symbolic links from your dotfiles directory to your home directory, making it easy to keep your dotfiles in a version-controlled repository.

## Why not GNU Stow?

While GNU Stow is a popular tool for managing dotfiles, it has several limitations that motivated the creation of this project:

| Stow Limitation | Our Solution |
|---|---|
| **Directory symlinks** - Stow creates directory-level symlinks, which can lead to software modifications polluting your dotfile repository | **File-only linking** - Only creates file-level symbolic links, preventing accidental pollution of your dotfiles |
| **No environment-specific configs** - Cannot handle different configurations for different environments (Linux/macOS/work/home) | **Template system** - Supports environment-specific configurations through templating and conditional deployment |
| **Unreliable uninstallation** - `unstow` cannot accurately track which files were originally installed | **State tracking** - Maintains a manifest of installed files for reliable uninstallation and verification |

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

### Commands

#### `install`

The `install` subcommand creates symbolic links for your dotfiles in your home directory.

```bash
# Basic usage
dotman install

# With debug logging
dotman --debug install

# With custom dotfiles directory
dotman --dir /path/to/dotfiles install
```

#### `uninstall`

The `uninstall` subcommand removes the symbolic links created by the `install` command.

```bash
dotman uninstall

# With debug mode
dotman --debug uninstall
```

#### `verify`

The `verify` subcommand checks if the symbolic links are correctly set up.

```bash
dotman verify

# With custom directory
dotman --dir /tmp/test-dotfiles verify
```

#### Getting Help

```bash
# Show main help
dotman --help

# Show help for a specific command
dotman install --help
```

## License

This project is licensed under the Apache-2 License.

# CLI Framework Specification

## ADDED Requirements

### Requirement: CLI Subcommand Structure
The application SHALL use Cobra framework to provide professional subcommand handling.

#### Scenario: Basic CLI Structure
```bash
# Root command shows help and available subcommands
$ dotman
A simple command-line tool to manage your dotfiles.

Usage:
  dotman [command]

Available Commands:
  install     Install dotfiles to home directory
  uninstall   Remove installed dotfiles
  verify      Verify dotfiles installation

Flags:
      --debug   Enable debug logging
  -h, --help    help for dotman
      --dir     Directory containing dotfiles (default "$HOME/.config/dotfiles")
```

#### Scenario: Subcommand Help
```bash
# Help for specific subcommand
$ dotman install --help
Install dotfiles to home directory

Usage:
  dotman install [flags]

Flags:
  -h, --help   help for install

Global Flags:
      --debug   Enable debug logging
      --dir     Directory containing dotfiles (default "$HOME/.config/dotfiles")
```

### Requirement: Command Registration
The application SHALL register install, uninstall, and verify subcommands under the root command.

#### Scenario: Subcommand Execution
```bash
# Execute install subcommand
$ dotman install
# Command executes install functionality (implementation defined in separate spec)

# Execute verify subcommand
$ dotman verify
# Command executes verification functionality (implementation defined in separate spec)
```

## MODIFIED Requirements

### Requirement: Application Entry Point
The application entry point SHALL be moved to `cmd/main.go` using Cobra command execution.

#### Scenario: Application Startup
```go
// cmd/main.go entry point
package main

import "github.com/spf13/cobra"

func main() {
    rootCmd := &cobra.Command{
        Use:   "dotman",
        Short: "A simple command-line tool to manage your dotfiles",
    }

    // Register subcommands and flags
    rootCmd.AddCommand(installCmd, uninstallCmd, verifyCmd)

    // Execute command system
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
```
# Command Flags Specification

## ADDED Requirements

### Requirement: Dotfiles Directory Flag
The application SHALL accept a `--dir` flag to specify the dotfiles directory location.

#### Scenario: Default Directory Behavior
```bash
# Uses default $HOME/.config/dotfiles
$ dotman install
{"level":"info","time":"2024-01-15T10:40:00Z","message":"Using dotfiles directory: /home/user/.config/dotfiles"}
```

#### Scenario: Custom Directory Flag
```bash
# Uses custom directory specified via flag
$ dotman install --dir /path/to/my/dotfiles
{"level":"info","time":"2024-01-15T10:40:00Z","message":"Using dotfiles directory: /path/to/my/dotfiles"}
```

#### Scenario: Short Flag Support
```bash
# Short flag variant
$ dotman install -d /path/to/my/dotfiles
{"level":"info","time":"2024-01-15T10:40:00Z","message":"Using dotfiles directory: /path/to/my/dotfiles"}
```

### Requirement: Debug Logging Flag
The application SHALL accept a `--debug` flag to enable debug-level logging.

#### Scenario: Debug Flag Usage
```bash
# Enable debug logging
$ dotman install --debug
{"level":"debug","time":"2024-01-15T10:40:00Z","message":"Debug mode enabled"}
{"level":"debug","time":"2024-01-15T10:40:01Z","message":"Detailed operation information"}
```

#### Scenario: Debug Flag with Other Options
```bash
# Combine debug with directory flag
$ dotman install --debug --dir /custom/path
{"level":"debug","time":"2024-01-15T10:40:00Z","message":"Debug mode enabled"}
{"level":"debug","time":"2024-01-15T10:40:00Z","message":"Using dotfiles directory: /custom/path"}
```

### Requirement: Persistent Global Flags
The `--dir` and `--debug` flags SHALL be available on all subcommands and inherited from root command.

#### Scenario: Flags Available on All Subcommands
```bash
# Directory flag works on install
$ dotman install --dir /custom/path

# Directory flag works on verify
$ dotman verify --dir /custom/path

# Directory flag works on uninstall
$ dotman uninstall --dir /custom/path

# Debug flag works on all subcommands
$ dotman verify --debug
$ dotman uninstall --debug
```

### Requirement: Flag Default Resolution
Directory resolution SHALL follow precedence: flag > environment variable > default.

#### Scenario: Environment Variable Fallback
```bash
# Set environment variable
export DOTFILES_DIR=/env/dotfiles

# Uses environment when no flag provided
$ dotman install
{"level":"info","time":"2024-01-15T10:40:00Z","message":"Using dotfiles directory: /env/dotfiles"}

# Flag overrides environment
$ dotman install --dir /flag/path
{"level":"info","time":"2024-01-15T10:40:00Z","message":"Using dotfiles directory: /flag/path"}
```

#### Scenario: Default Directory Resolution
```bash
# No flag or environment variable
$ dotman install
{"level":"info","time":"2024-01-15T10:40:00Z","message":"Using dotfiles directory: /home/user/.config/dotfiles"}
```

## MODIFIED Requirements

### Requirement: Root Command Configuration
The root command SHALL be enhanced to support persistent global flags.

#### Scenario: Root Command Flag Setup
```go
// cmd/root.go
package cmd

import (
    "os"
    "os/user"
    "github.com/spf13/cobra"
)

var (
    debugFlag bool
    dirFlag   string
)

func init() {
    // Persistent flags available to all subcommands
    rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
    rootCmd.PersistentFlags().StringVarP(&dirFlag, "dir", "d", "", "Directory containing dotfiles")

    // Set up directory resolution with defaults
    cobra.OnInitialize(initConfig)
}

func initConfig() {
    if dirFlag == "" {
        if envDir := os.Getenv("DOTFILES_DIR"); envDir != "" {
            dirFlag = envDir
        } else {
            if usr, err := user.Current(); err == nil {
                dirFlag = filepath.Join(usr.HomeDir, ".config", "dotfiles")
            }
        }
    }

    // Initialize logger with debug flag
    logger.Init(debugFlag)
}
```
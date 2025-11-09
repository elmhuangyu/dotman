# Logging System Specification

## ADDED Requirements

### Requirement: Structured Logging with Zerolog
The application SHALL use zerolog for structured logging with configurable levels.

#### Scenario: Default Info Level Logging
```bash
# Default operation with info-level logging
$ dotman install
{"level":"info","time":"2024-01-15T10:30:00Z","message":"Starting dotfiles installation"}
{"level":"info","time":"2024-01-15T10:30:01Z","message":"Successfully installed 15 dotfiles"}
```

#### Scenario: Debug Level Logging
```bash
# With --debug flag enabled
$ dotman install --debug
{"level":"debug","time":"2024-01-15T10:30:00Z","message":"Initializing logger with debug level"}
{"level":"debug","time":"2024-01-15T10:30:00Z","message":"Dotfiles directory: /home/user/.config/dotfiles"}
{"level":"info","time":"2024-01-15T10:30:00Z","message":"Starting dotfiles installation"}
{"level":"debug","time":"2024-01-15T10:30:00Z","message":"Processing config.yaml"}
{"level":"debug","time":"2024-01-15T10:30:01Z","message":"Creating symlink: .vimrc -> /home/user/.config/dotfiles/vimrc"}
{"level":"info","time":"2024-01-15T10:30:01Z","message":"Successfully installed 15 dotfiles"}
```

### Requirement: Global Logger Configuration
The application SHALL configure a global logger instance accessible throughout the application.

#### Scenario: Logger Initialization
```go
// pkg/logger/logger.go
package logger

import (
    "os"
    "github.com/rs/zerolog"
)

var Logger zerolog.Logger

func Init(debug bool) {
    output := zerolog.ConsoleWriter{Out: os.Stderr}

    if debug {
        Logger = zerolog.New(output).With().Timestamp().Logger().Level(zerolog.DebugLevel)
    } else {
        Logger = zerolog.New(output).With().Timestamp().Logger().Level(zerolog.InfoLevel)
    }
}
```

#### Scenario: Logger Usage in Commands
```go
// cmd/commands/install.go
package commands

import "github.com/elmhuangyu/dotman/pkg/logger"

func runInstall(cmd *cobra.Command, args []string) {
    logger.Logger.Info().Msg("Starting dotfiles installation")

    // Operation details...

    logger.Logger.Debug().Str("file", ".vimrc").Msg("Processing dotfile")
    logger.Logger.Info().Int("count", 15).Msg("Successfully installed dotfiles")
}
```

### Requirement: Debug Flag Integration
The `--debug` flag SHALL control the global log level across all subcommands.

#### Scenario: Debug Flag Affects All Commands
```bash
# Debug logging works for all subcommands
$ dotman verify --debug
{"level":"debug","time":"2024-01-15T10:35:00Z","message":"Initializing logger with debug level"}
{"level":"debug","time":"2024-01-15T10:35:00Z","message":"Verification mode: checking all symlinks"}
{"level":"info","time":"2024-01-15T10:35:01Z","message":"All dotfiles verified successfully"}
```

## ADDED Dependencies

- `github.com/rs/zerolog` - Structured logging library
- `github.com/rs/zerolog/log` - Optional convenience functions
# Implement CLI Framework with Cobra and Logging

## Summary

This change adds the Cobra CLI framework for professional subcommand handling, integrates zerolog for structured logging with configurable levels, and adds essential command-line flags including `--dir` for dotfiles directory configuration and `--debug` for verbose logging.

## Problem

The current dotman project lacks a proper CLI structure and logging infrastructure. As a new Go project, it needs:
- Professional subcommand handling for install/uninstall/verify operations
- Configurable logging for debugging and operational visibility
- Flexible configuration via command-line flags
- A proper entry point following Go CLI best practices

## Solution

Implement a comprehensive CLI foundation using:
- **Cobra** - The de-facto standard for CLI applications in Go
- **Zerolog** - Zero-allocation structured logging
- **Command flags** - `--dir` for dotfiles directory, `--debug` for log level control
- **cmd/main.go** - Proper application entry point

## Capabilities

1. **CLI Framework** - Professional subcommand structure with Cobra
2. **Logging System** - Structured logging with configurable levels
3. **Configuration Flags** - Essential command-line options for flexibility

## Impact

This establishes the foundation for all future CLI functionality and provides a professional user experience for dotfile management operations.

## Why

The current dotman project lacks a proper CLI structure and logging infrastructure. As a new Go project, it needs a professional foundation for:
- Professional subcommand handling for install/uninstall/verify operations
- Configurable logging for debugging and operational visibility
- Flexible configuration via command-line flags
- A proper entry point following Go CLI best practices

Without this foundation, the tool would have poor user experience, limited debugging capabilities, and would not follow Go CLI conventions.

## What Changes

This implementation adds:

1. **Dependencies**: Cobra CLI framework and Zerolog logging library to go.mod
2. **Directory Structure**:
   - `cmd/main.go` - Application entry point
   - `cmd/` - Root command and subcommands
   - `pkg/logger/` - Centralized logging configuration
3. **Core Features**:
   - Root command with global flags (--debug, --dir)
   - Three subcommands: install, uninstall, verify (stub implementations)
   - Structured logging with configurable levels
   - Directory resolution with sensible defaults
4. **Documentation**: Updated README.md with CLI usage examples

## Dependencies

- Go 1.25.4 (already specified in go.mod)
- No external system dependencies
## Context
The dotman tool needs to support modular dotfile management where each module (e.g., nvim, git, bash) has its own directory with configuration files that need to be mapped to specific locations in the user's system.

## Goals / Non-Goals
- Goals:
  - Enable per-module configuration for file destinations
  - Support simple YAML configuration format
  - Maintain backward compatibility with existing flag-based directory specification
- Non-Goals:
  - Complex templating or variable substitution in this iteration
  - Multiple target directories per module (keep it simple initially)

## Decisions
- Decision: Use YAML format for "Dotfile" configuration
- Rationale: YAML is human-readable, widely used, and easy to parse in Go
- Alternatives considered: JSON (less readable), TOML (less common for this use case)

- Decision: Place "Dotfile" in module directories
- Rationale: Each module should be self-contained with its own configuration
- Alternatives considered: Central config file (would be more complex to manage)

- Decision: Single `target_dir` field for now
- Rationale: Simple implementation that covers the most common use case
- Future extensions: Could add file-specific mappings, ignore patterns, etc.

## Risks / Trade-offs
- Risk: Users might expect more complex configuration options
- Mitigation: Document current limitations clearly, design for future extensibility
- Trade-off: Simplicity vs. flexibility - starting simple but with extension points

## Migration Plan
- No migration needed - this is additive functionality
- Existing flag-based usage continues to work unchanged
- New config file usage is optional initially

## Open Questions
- Should the config support relative or absolute paths for target_dir?
- How should conflicts be handled if both --dir flag and config file exist? (Current thinking: flag should override)
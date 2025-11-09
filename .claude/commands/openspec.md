---
name: OpenSpec
description: OpenSpec spec-driven development CLI - main entry point
category: OpenSpec
tags: [openspec, cli, main]
---
<!-- OPENSPEC:START -->
**Guardrails**
- Favor straightforward, minimal implementations first and add complexity only when it is requested or clearly required.
- Keep changes tightly scoped to the requested outcome.
- Refer to `openspec/AGENTS.md` (located inside the `openspec/` directory—run `ls openspec` or `openspec update` if you don't see it) if you need additional OpenSpec conventions or clarifications.

**Main Entry Point**
This is the root OpenSpec command that provides access to all OpenSpec functionality. It serves as the main CLI interface for spec-driven development.

**Available Subcommands**
- `/openspec:proposal` - Create new change proposals
- `/openspec:apply` - Implement approved changes
- `/openspec:archive` - Archive deployed changes

**Quick Start Workflow**
1. **List current state**: Run `openspec list` to see active changes and `openspec list --specs` to see existing specifications
2. **Create proposal**: Use `/openspec:proposal` to scaffold new change proposals when adding features or making breaking changes
3. **Implement changes**: Use `/openspec:apply` to implement approved proposals
4. **Archive completed work**: Use `/openspec:archive` to move deployed changes to archive

**Common Operations**
```bash
# Status checks
openspec list                  # Show active changes
openspec list --specs          # Show current specifications
openspec show <item>           # View details of changes or specs
openspec validate <item>       # Validate changes or specs

# Change management
/openspec:proposal             # Create new change proposal
/openspec:apply <change-id>    # Implement approved change
/openspec:archive <change-id>  # Archive completed change

# Project management
openspec init [path]           # Initialize OpenSpec in new project
openspec update [path]         # Update instruction files
```

**When to Use Each Command**
- **Planning/Design phase**: Start with `/openspec:proposal` for new features, breaking changes, or architecture shifts
- **Implementation phase**: Use `/openspec:apply` after proposal approval
- **Deployment phase**: Use `/openspec:archive` after changes are deployed
- **Status/Review**: Use `openspec list`, `openspec show`, and `openspec validate` throughout development

**Reference**
- Complete documentation available in `openspec/AGENTS.md`
- Use `openspec show <item> --json` for programmatic access
- Run `openspec validate --strict` for comprehensive validation
<!-- OPENSPEC:END -->
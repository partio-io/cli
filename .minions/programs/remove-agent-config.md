---
id: remove-agent-config
target_repos:
  - cli
acceptance_criteria:
  - "`partio disable --agent <name>` removes configuration for a specific agent without disabling partio entirely"
  - Agent-specific hook files and config entries are cleaned up
  - Other enabled agents remain unaffected
  - Running without `--agent` flag preserves existing behavior (full disable)
  - The command prints a summary of what was removed
pr_labels:
  - minion
---

# Support removing individual agent configuration

## Description

Add the ability to remove configuration for a specific agent without fully disabling Partio. Currently, `partio disable` is all-or-nothing — it removes hooks and the `.partio/` directory entirely. Users who have multiple agents configured (or who want to switch agents) need a way to cleanly remove one agent's configuration while keeping Partio active for others.

## Why

As Partio adds support for more agents beyond Claude Code, users will need to manage agent configurations independently. Switching from one agent to another currently requires a full disable/enable cycle, which resets all configuration.

## Context hints

- `cmd/partio/disable.go` — disable command
- `cmd/partio/enable.go` — enable command (for reference on agent-specific setup)
- `internal/config/` — configuration management

## Source

Inspired by entireio/cli changelog 0.5.4 (`configure --remove-agent` functionality).

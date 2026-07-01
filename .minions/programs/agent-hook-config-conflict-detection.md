---
id: agent-hook-config-conflict-detection
target_repos:
  - cli
acceptance_criteria:
  - During `partio enable`, validate that generated agent hook config files do not conflict with the agent's own default configuration
  - Warn the user when a configuration file Partio creates would trigger warnings or unexpected behavior in the target agent
  - Support a --force flag to proceed despite detected conflicts
  - Add conflict detection rules for Claude Code hook configuration
pr_labels:
  - minion
---

# Detect agent hook configuration conflicts during setup

## Summary

When `partio enable` configures hooks for an agent, validate that the resulting configuration files do not conflict with the agent's own defaults or trigger unexpected warnings. Warn users about detected conflicts before writing configuration.

## Why

Agents have their own configuration expectations and validation rules. When Partio writes hook configuration files for an agent (e.g., Claude Code's hooks in `.claude/settings.local.json`), it may inadvertently create entries that conflict with the agent's defaults, trigger deprecation warnings, or cause the agent to behave unexpectedly. Users currently discover these conflicts only after the agent shows warnings, leading to confusion about whether Partio or the agent is misconfigured. Proactive conflict detection during `partio enable` prevents this class of issues.

## What to implement

1. Add a `ValidateConfig` method to the `agent.Detector` interface (or as an optional interface agents can implement).
2. In the Claude Code detector, implement validation rules that check for known conflict patterns (e.g., hook paths that shadow agent defaults, deprecated config keys, configuration entries that trigger agent-side warnings).
3. During `partio enable --agent <name>`, call the validator before writing config files.
4. If conflicts are detected, display warnings describing each conflict and what the user should do. Proceed only with `--force` or user confirmation.
5. Run the same validation in `partio doctor` to catch conflicts introduced after initial setup.

## Context hints

- `internal/agent/detector.go` — Detector interface
- `internal/agent/claude/` — Claude Code agent implementation
- `cmd/partio/enable.go` — enable command
- `cmd/partio/doctor.go` — doctor command

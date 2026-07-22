---
id: enable-auto-import-non-interactive
target_repos:
  - cli
acceptance_criteria:
  - "`partio enable --yes` automatically imports all pre-existing Claude Code sessions without prompting"
  - "When no TTY is detected (non-interactive context), `partio enable` auto-imports without prompting"
  - "Import runs after hooks are installed, before reporting success"
  - "If no importable sessions exist, the step is silently skipped"
  - "Import errors are logged as warnings; they do not cause `partio enable` to exit non-zero"
  - "Interactive `partio enable` (no `--yes`, has TTY) behavior is unchanged"
pr_labels:
  - minion
  - enhancement
---

# Auto-import pre-existing sessions when `partio enable` runs non-interactively

## Problem

`partio enable` installs hooks and writes config, but does nothing about the Claude Code session history that already exists on disk. Users who enable Partio mid-project miss all context from prior sessions unless they discover and run `partio import` separately. In non-interactive contexts (CI, scripted onboarding, `--yes` flag), there is no opportunity to prompt for this.

## Proposed behavior

When `partio enable` is run in a non-interactive context — either because `--yes` is passed, or because there is no TTY — it should automatically import all eligible pre-existing Claude Code sessions after installing hooks. This matches the behavior introduced in entireio/cli 0.8.1, where first-time enable auto-imports all eligible agents when run non-interactively.

Concretely:
- After hooks are installed, call the same logic as `partio import` to scan `~/.claude/projects/` for sessions tied to the current repo.
- Run the import silently (no prompts); log a summary line of what was imported (`partio: imported N session(s) from prior history`).
- If no sessions are found, skip silently.
- Import failures are non-fatal; log a warning and continue.

Interactive `partio enable` (TTY present, no `--yes`) is out of scope for this change — a separate follow-up can add an interactive prompt.

## Why this matters

The first `git commit` after `partio enable` is rarely the first commit in a project. Without auto-import, the checkpoint branch starts empty even though there may be weeks of Claude Code session history available. Auto-importing on enable closes this gap for CI/scripted onboarding with zero extra steps.

## Source

entireio/cli changelog 0.8.1 — "First-time `entire enable` now offers to import pre-existing agent history for the agents you select; non-interactive runs (`--yes` or no TTY) auto-import all eligible agents."

## Context hints

- `cmd/partio/` — enable command entry point
- `internal/agent/claude/` — session discovery logic used by import
- `internal/checkpoint/` — checkpoint write path

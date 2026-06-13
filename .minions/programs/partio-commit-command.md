---
id: partio-commit-command
target_repos:
  - cli
acceptance_criteria:
  - "`partio commit -m <message>` creates a git commit with checkpoint metadata embedded as a git header"
  - "The command detects the active agent session and prompts the user to link it"
  - "Checkpoint trailer is written both as a git header and as a commit message trailer for backwards compatibility"
  - "The command works without requiring git hooks to be installed (hookless workflow)"
  - "Attribution is calculated and stored in checkpoint metadata during the commit"
  - "Session transcript is condensed after checkpoint creation"
  - "All flags supported by `git commit` that are relevant (e.g., `-a`, `--amend`, `--allow-empty`) are passed through"
  - "Unit tests cover the commit object creation and header embedding"
pr_labels:
  - minion
---

# Add `partio commit` command for hookless checkpoint capture

## Description

Add a `partio commit` command that wraps `git commit` to directly embed checkpoint information as a git header in the commit object. This provides a hookless alternative to the current git-hooks-based workflow, giving users explicit control over when checkpoint metadata is attached.

The command should:
1. Stage and create the commit (delegating to git or using go-git)
2. Detect the active agent session (Claude Code, etc.)
3. Prompt the user to confirm linking the session to the commit
4. Calculate attribution and create the checkpoint
5. Write the checkpoint ID both as a custom git header (`checkpoint <id>`) and as an `Partio-Checkpoint` trailer in the commit message
6. Condense the session transcript

This enables users who cannot or prefer not to install git hooks (e.g., in environments with hook managers, CI systems, or shared repositories with hook policies) to still capture checkpoint data with every commit.

## Why

The current hook-based approach requires `partio enable` to install git hooks, which can conflict with other hook managers (Husky, Lefthook, hk) or be stripped by CI environments. A dedicated commit command provides a reliable, explicit alternative that doesn't depend on hook infrastructure. It also gives users clearer visibility into what Partio is doing at commit time.

## Source

entireio/cli PR #924 — `entire commit` command that embeds checkpoint info as git headers

## Context hints

- `cmd/partio/` — CLI command definitions
- `internal/checkpoint/` — checkpoint creation and storage
- `internal/hooks/` — existing hook implementations (for reference on attribution + checkpoint flow)
- `internal/session/` — session detection and management

---
id: worktree-settings-resolution
target_repos:
  - cli
acceptance_criteria:
  - When running in a git worktree, partio resolves .partio/settings.json from the main working tree root
  - .partio/settings.local.json is resolved from both the worktree root and main working tree root, with worktree-local taking precedence
  - partio enable in a worktree does not create a duplicate .partio/ directory in the worktree
  - partio status correctly reports configuration when run from a worktree
  - Existing single-repo (non-worktree) behavior is unchanged
pr_labels:
  - minion
---

# Resolve Partio settings from main working tree in git worktrees

## Summary

When Partio runs inside a git worktree, settings resolution should look for `.partio/settings.json` in the main working tree root (not the worktree root), matching how hooks are already installed to `git rev-parse --git-common-dir`.

## Motivation

Partio already supports git worktrees for hook installation and session discovery, but settings resolution does not account for worktree layouts. If a user runs `partio enable` in the main working tree and then creates a worktree, commands run in the worktree may not find the `.partio/` directory because it only exists in the main tree. This causes `partio status` to report "not enabled" and hooks to silently skip checkpoint creation.

Inspired by entireio/cli#1159 which fixed dropped checkpoints when agents run from linked worktrees by aligning transcript, settings, and session lookup paths.

## Design

1. **Settings resolution** (`internal/config/`): When loading settings, first check the current repo root for `.partio/settings.json`. If not found and the current directory is a worktree (detected via `git rev-parse --git-common-dir` differing from `--git-dir`), resolve the main working tree root via `git worktree list --porcelain` and check there.

2. **Local settings layering**: `.partio/settings.local.json` should be checked in both locations, with the worktree-local file taking precedence over the main tree's local settings. This allows per-worktree overrides.

3. **Enable command**: When `partio enable` is run in a worktree, detect this and either configure from the main tree or warn the user to run enable from the main working tree.

4. **Status command**: Report the resolved settings path so users can see which settings file is being used.

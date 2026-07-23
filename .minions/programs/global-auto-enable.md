---
id: global-auto-enable
target_repos:
  - cli
acceptance_criteria:
  - "A new `auto_enable` boolean setting is supported in the global config (`~/.config/partio/settings.json`)"
  - "When `auto_enable: true`, the post-commit hook silently skips checkpoint creation (or logs a debug message) rather than erroring when Partio is not explicitly enabled in the repo"
  - "Alternatively: when `auto_enable: true`, a detected active agent session in an unenabled repo triggers automatic hook install on the fly"
  - "`partio enable --global` sets `auto_enable: true` in the global config"
  - "Repos that have explicitly run `partio disable` are excluded from auto-enable"
  - "`make test` passes"
  - "`make lint` passes"
pr_labels:
  - minion
  - enhancement
---

# Global auto-enable: track agent sessions without per-repo configuration

## Background

Inspired by entireio/cli#1098.

Currently, Partio requires running `partio enable` in each repository before it captures any sessions. Users who frequently create new repos, work across many projects, or prototype quickly often miss enabling Partio and lose session context as a result.

## What to implement

Add a global `auto_enable` configuration option that allows Partio to capture agent sessions in any git repo without requiring per-repo setup.

Two possible approaches (implement whichever fits the architecture better):

**Option A — Silent global hooks**: Install Partio as a global git hook template (`~/.config/git/hooks/` or `init.templateDir`) when `partio enable --global` is run. Global hooks check for an active agent session and create checkpoints, but don't write any per-repo config. Repos that have run `partio disable` set a `PARTIO_ENABLED=false` override to opt out.

**Option B — Global config flag**: Add `auto_enable: true` to the global config. The pre-commit/post-commit hooks (when installed globally or triggered from PATH) detect the flag and proceed with checkpoint creation even when the current repo has no Partio config.

In either case:
- Expose `partio enable --global` as the user-facing command to activate this mode
- Respect the layered config: repo-level `enabled: false` takes precedence over global auto-enable
- Add `--global` to `partio disable` to turn it off globally

## Why this matters

The per-repo enable requirement creates friction that causes users to miss capturing valuable session context. Global auto-enable makes Partio "just work" for power users who move between many repos, especially those creating new projects frequently.

## Context hints

- `internal/config/` (layered config system)
- `cmd/partio/` (enable, disable commands)
- `internal/hooks/` (hook implementations that read config)

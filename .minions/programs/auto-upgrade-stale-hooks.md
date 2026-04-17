---
id: auto-upgrade-stale-hooks
target_repos:
  - cli
acceptance_criteria:
  - "`partio enable` compares installed hook content against expected content and updates hooks that differ"
  - "Hooks that already match expected content are left untouched (no unnecessary writes)"
  - "Backup chain is preserved — existing `.partio-backup` files are not overwritten during upgrade"
  - "A message is printed when hooks are upgraded (e.g., 'Updated pre-commit hook')"
  - "The `.partio/` directory existence check no longer short-circuits the entire enable flow"
  - "Table-driven tests cover: fresh install, stale hook upgrade, already-current hook skip, and backup preservation"
pr_labels:
  - minion
---

# Auto-upgrade stale hooks on `partio enable`

## Problem

When a user upgrades the Partio CLI binary, the git hook scripts installed in their repositories become stale — they may reference old binary paths, use outdated shim logic, or miss new hooks added in later versions. Currently, `partio enable` short-circuits entirely when `.partio/` exists (`cmd/partio/enable.go:38-42`), telling the user "partio is already enabled" without checking whether the installed hooks are up to date.

The only workaround is `partio disable && partio enable`, which users must know to do and which unnecessarily tears down and rebuilds state.

## Desired behavior

`partio enable` should detect stale hooks and silently upgrade them:

1. If `.partio/` exists, skip directory/config creation but still check hooks.
2. For each expected hook (`pre-commit`, `post-commit`, `pre-push`), read the installed script from the hooks directory (`git rev-parse --git-common-dir`).
3. Compare the installed content against the expected content from `hookScript()` / `hookScriptAbsolute()` (`internal/git/hooks/hooks.go`).
4. If the content differs, overwrite the hook file (preserving the existing backup chain — do not re-backup a partio-managed hook).
5. If a hook is missing entirely (e.g., a new hook type added in a later version), install it following the normal backup logic.
6. Print a summary of what was upgraded, or "partio is already enabled and hooks are up to date" if nothing changed.

## Context hints

- `cmd/partio/enable.go` — the `.partio/` existence check that needs to be relaxed
- `internal/git/hooks/install.go` — `installHooks()` needs an upgrade-aware code path
- `internal/git/hooks/hooks.go` — `hookScript()` / `hookScriptAbsolute()` generate expected content
- `internal/git/hooks/uninstall.go` — backup restoration logic to preserve during upgrades

## Source

Inspired by entireio/cli PR #775 — OpenCode plugin hook rewrite when content differs.

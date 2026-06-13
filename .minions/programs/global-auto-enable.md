---
id: global-auto-enable
target_repos:
  - cli
acceptance_criteria:
  - "`partio enable --global` installs user-level git hooks via `git config --global core.hooksPath` or init.templateDir"
  - "New git repos automatically have partio hooks without running `partio enable` per-repo"
  - "`partio disable --global` removes user-level hooks and restores previous hooksPath"
  - "Per-repo `partio disable` overrides global enable for that specific repo"
  - "Global mode is stored in the global config file (~/.config/partio/settings.json)"
pr_labels:
  - minion
  - enhancement
---

# Add global auto-enable for automatic session tracking

## Description

Add a `partio enable --global` flag that installs user-level git hooks so that every git repository on the machine automatically tracks agent sessions without requiring per-repo `partio enable`.

Currently, Partio requires explicit per-repo enablement. Users who create repos frequently (rapid prototyping, POCs) miss capturing their most interesting agent sessions because they forget to run `partio enable` in each new repo.

## Implementation Notes

- Add `--global` flag to `partio enable` that configures hooks at the user level
- Consider using `git config --global core.hooksPath` pointing to a managed hooks directory, or `init.templateDir` for new repos
- The managed hooks directory should chain to any existing user hooks
- Per-repo disable should take precedence over global enable (check `.partio/settings.json` for `enabled: false`)
- `partio status` should indicate whether tracking is active due to global or per-repo config
- `partio disable --global` should cleanly restore previous `core.hooksPath` if one existed

## Context Hints

- `cmd/partio/enable.go` — current enable command
- `cmd/partio/disable.go` — current disable command
- `internal/git/hooks/` — hook installation logic
- `internal/config/` — layered configuration system

---
id: checkpoint-sync-destination
target_repos:
  - cli
acceptance_criteria:
  - "partio enable presents an interactive prompt to choose checkpoint sync destination when running in a TTY"
  - "Supported destinations include in-repo orphan branch (default) and external git repository URL"
  - "Selected destination is persisted in .partio/settings.json as a checkpoint_remote field"
  - "pre-push hook reads checkpoint_remote and pushes checkpoint branch to the configured remote instead of origin when set"
  - "partio status displays the configured checkpoint sync destination"
  - "Non-interactive mode (CI/pipes) skips the picker and uses the default"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add checkpoint sync destination picker to partio enable

## Summary

Add an interactive step to `partio enable` that lets users choose where their checkpoint data is synced. Currently checkpoints are always pushed to origin alongside the main repository. Users should be able to select an alternative destination (e.g., a dedicated private repository) to keep session data separate from the source code remote.

## Context

- `internal/config/` — layered configuration system
- `internal/git/` — git operations, remote management
- `cmd/partio/enable.go` — enable command implementation
- `internal/hooks/pre_push.go` — pre-push hook that pushes checkpoint branch

## Implementation notes

- Add a `checkpoint_remote` field to the Config type
- During `partio enable`, if running interactively, prompt the user to choose: "Store checkpoints in this repository (default)" or "Store checkpoints in a separate repository" with a URL input
- Update the pre-push hook to read `checkpoint_remote` from config and use it as the push target
- Show the configured destination in `partio status` output

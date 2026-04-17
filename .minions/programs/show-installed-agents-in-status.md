---
id: show-installed-agents-in-status
target_repos:
  - cli
acceptance_criteria:
  - "`partio status` shows a line listing which agents have hooks installed when partio is enabled"
  - "Detection is based on whether hook files exist and contain the partio marker, not just what is in config"
  - "Line is omitted when partio is disabled or no agent hooks are detected"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Show agents with hooks installed in `partio status`

Currently `partio status` shows `Agent: claude-code` from the config file, but this reflects configuration intent rather than the actual installed state. Users who edit configs or run `partio enable` multiple times can end up with a mismatch between config and reality.

Add an "Agents with hooks:" line below `Hooks:` in the status output that shows which agents actually have hooks installed on disk. Detection should read the existing hook files and check for the partio marker (already defined in `internal/git/hooks/hooks.go` as `partioMarker = "# Installed by partio"`).

## Desired output

```
Status:     enabled
Strategy:   manual-commit
Agent:      claude-code
Hooks:      installed
Agents:     claude-code
Checkpoints: branch exists
```

If no partio-owned hook files are found, omit the "Agents:" line. Partio currently only supports claude-code, so for now the line will show `claude-code` when hooks are installed and be omitted otherwise.

## Key files

- `cmd/partio/status.go` — `runStatus()` — add the new output line after the hooks check
- `internal/git/hooks/hooks.go` — `isPartioHook()` helper for checking hook content
- `internal/git/hooks/` — `hookNames` slice lists the expected hook names

**Inspired by:** entireio/cli#847 (Show installed agents in status output)

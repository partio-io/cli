---
id: hooks-local-settings-activation-guard
target_repos:
  - cli
acceptance_criteria:
  - "Running `partio enable --local` in a fresh repo (no .partio/settings.json) installs hooks that create checkpoints on commit"
  - "Hook activation check considers settings.local.json as a valid activation source when settings.json is absent"
  - "Existing behavior is preserved when both settings.json and settings.local.json exist"
  - "partio doctor reports a warning if hooks are installed but no settings file is found"
pr_labels:
  - minion
---

# Ensure hooks activate when only local settings exist

## Description

When a user runs `partio enable --local`, hooks are installed but only `.partio/settings.local.json` is created. If the hook activation check requires `.partio/settings.json` to exist, hooks silently no-op — no checkpoints are saved, no errors are logged.

Update the hook activation logic to treat `.partio/settings.local.json` as a valid activation source. The layered config system already merges local settings, but the hook entry-point guard may short-circuit before reaching the config merge.

## Why

Users who prefer local-only configuration (to avoid committing settings to the repo) get a silently broken experience. The hooks appear to run but produce no output, making the failure very difficult to diagnose.

## Context hints

- `internal/hooks/` — hook implementations with activation guards
- `internal/config/` — layered configuration loading
- `cmd/partio/` — enable command creating settings files

---
id: auto-recover-hooks-after-external-overwrite
target_repos:
  - cli
acceptance_criteria:
  - When an external hook manager (Lefthook, Husky) overwrites a Partio-managed hook script, the next Partio hook invocation detects the overwrite and warns the user
  - partio doctor detects when hook scripts have been overwritten by another tool and reports the conflict
  - A recovery mechanism restores Partio hook functionality without requiring full disable/enable cycle
  - Hook backup chaining is preserved so both Partio and the external manager's hooks run
pr_labels:
  - minion
  - minion-proposal
---

# Auto-recover hooks after external hook manager overwrites

## Problem

Partio installs git hooks (pre-commit, post-commit, pre-push) to `.git/hooks/` (via `git rev-parse --git-common-dir`). When a repository also uses an external hook manager like Lefthook, Husky, or Rush, those tools may silently reclaim the hook files on their next config change or install step. This causes Partio's session capture and checkpoint push to stop working with no warning to the user.

The current workaround is `partio disable && partio enable`, but users may not notice the breakage for days or weeks.

## Proposed solution

1. **Hook integrity check**: Add a lightweight self-check at the top of each Partio hook script that verifies the script still contains the Partio marker comment. If the marker is missing, emit a warning to stderr and skip Partio-specific logic gracefully.

2. **Doctor detection**: Extend `partio doctor` to check whether each managed hook script still contains the Partio marker. If overwritten, report which hook manager likely replaced it (by checking for Lefthook/Husky/Rush markers) and suggest recovery steps.

3. **Recovery command**: Add `partio hooks repair` (or extend `partio enable --repair`) that re-injects Partio's hook logic into the current hook scripts without overwriting the external manager's content. This preserves the chain so both tools' hooks execute.

4. **Lefthook/Husky wrapper detection**: When `partio enable` or `partio doctor` detects a `.lefthook.yml`, `.husky/` directory, or similar config, proactively warn about potential hook conflicts and suggest using the external hooks backend (if available) or the repair workflow.

## Context hints

- `internal/git/hooks/` - Hook script generation and install/uninstall
- `cmd/partio/doctor.go` - Doctor command implementation
- `cmd/partio/enable.go` - Enable command implementation

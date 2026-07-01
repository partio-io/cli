---
id: graceful-hook-degradation
target_repos:
  - cli
acceptance_criteria:
  - "Hook scripts check if the partio binary exists before invoking it"
  - "When partio binary is not found, hooks exit silently with success (exit 0)"
  - "Git operations (commit, push) are never blocked by a missing partio binary"
  - "A one-line warning is printed to stderr when the binary is missing so users know hooks are stale"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Graceful hook degradation when partio binary is missing

## Problem

When `partio` is uninstalled but repo-level git hooks still reference the `partio _hook` command, git operations surface noisy failures. Agent tools like Claude Code and Codex may display confusing error messages from the hook scripts, and users see unexpected errors on every commit or push.

## Solution

Update the generated hook scripts (in `internal/git/hooks/`) to check whether the `partio` binary is available (e.g., `command -v partio`) before attempting to invoke it. If the binary is not found:
- Print a short warning to stderr: `partio: binary not found, skipping hook (run 'partio disable' to remove hooks)`
- Exit with code 0 so git operations proceed normally

## Context

- Inspired by entireio/cli#880
- Relevant code: `internal/git/hooks/` (hook script generation)

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```

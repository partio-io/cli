---
id: validate-agent-binary-on-enable
target_repos:
  - cli
acceptance_criteria:
  - "partio enable warns if the partio binary is not found in the PATH that hook scripts will use at execution time"
  - "The warning includes actionable guidance on how to fix the PATH issue"
  - "Hook installation still proceeds despite the warning (non-blocking)"
  - "partio doctor also checks for this condition and reports it"
pr_labels:
  - minion
---

# Validate agent binary availability on enable

When `partio enable` installs git hooks, those hooks invoke the `partio` binary during git operations. If the binary isn't in PATH at hook execution time (e.g., installed via `go install` but GOPATH/bin not in the shell's PATH, or installed in a non-standard location), hooks silently fail or produce confusing errors.

## What to implement

1. During `partio enable`, after hook installation, verify that `partio` (or the binary that's currently running) would be resolvable from a fresh shell context similar to how git hooks execute.

2. If the binary can't be found, emit a warning like:
   ```
   Warning: partio binary may not be accessible from git hooks.
   Hooks are installed but may fail at commit time.
   Ensure the partio binary is in your PATH or use an absolute path in hook scripts.
   ```

3. Add a corresponding check to `partio doctor` that validates hook binary accessibility.

## Context hints

- `cmd/partio/` - CLI command implementations
- `internal/git/hooks/` - Hook script generation and installation

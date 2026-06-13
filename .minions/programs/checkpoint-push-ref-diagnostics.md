---
id: checkpoint-push-ref-diagnostics
target_repos:
  - cli
acceptance_criteria:
  - When checkpoint push fails, the specific ref rejection reason is surfaced to the user (e.g., "non-fast-forward", "permission denied")
  - Generic "push failed" errors are replaced with actionable messages that include the ref name and rejection reason
  - Suggested remediation is included (e.g., "run partio doctor" or "fetch and retry")
  - Non-zero exit from git push is still handled gracefully (hook does not block the user's push)
pr_labels:
  - minion
  - dx
---

# Surface per-ref rejection reasons on checkpoint push failure

## Problem

When `pre-push` hook pushes the checkpoint branch and it fails, the error message is generic (e.g., "failed to push checkpoint branch"). The actual rejection reason from the remote (non-fast-forward, permission denied, branch protection, etc.) is not surfaced. Users see a warning but have no actionable information about what went wrong or how to fix it.

## Proposed Solution

Parse the git push stderr/stdout for per-ref status lines and surface them:

1. Capture both stdout and stderr from the `git push` subprocess for the checkpoint branch.
2. Parse ref-status lines (git push reports per-ref results like `! [rejected] ... (non-fast-forward)`).
3. Surface the specific reason in the warning message shown to the user.
4. Include a remediation hint based on the rejection type:
   - Non-fast-forward → "Run `partio doctor` to reconcile"
   - Permission denied → "Check remote permissions for the checkpoint branch"
   - Unknown → "Run with PARTIO_LOG_LEVEL=debug for full output"

## Context

- `internal/hooks/` — pre-push hook implementation
- `internal/git/` — git push operations

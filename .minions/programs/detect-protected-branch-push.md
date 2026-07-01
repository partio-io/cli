---
id: detect-protected-branch-push
target_repos:
  - cli
acceptance_criteria:
  - Pre-push hook detects when the checkpoint branch ref targets a protected branch and skips the push with a warning instead of failing silently or producing a confusing git error
  - Detection uses the GitHub API (via gh cli) or parses the git push error to identify protection rule failures
  - A clear warning message is logged explaining that checkpoint push was skipped due to branch protection, with guidance on configuring an alternative checkpoint remote
  - Existing behavior is preserved when the checkpoint branch is not protected (push proceeds normally)
  - Tests cover both protected and unprotected branch scenarios
pr_labels:
  - minion
---

# Detect and handle protected branch push failures for checkpoint refs

## Problem

When the checkpoint branch (`partio/checkpoints/v1`) is pushed to a remote that has branch protection rules, the push can fail with a confusing git error. The current pre-push hook catches the error and logs a generic warning ("could not push checkpoint branch"), but doesn't help the user understand *why* or what to do about it.

This is especially problematic in organizations that use broad branch protection patterns (e.g., `*` or `partio/*`) that inadvertently cover checkpoint refs.

## Proposed Solution

Add detection in the pre-push hook to identify when a checkpoint push fails due to branch protection, and provide an actionable warning message. Two approaches:

1. **Reactive**: Parse the git push error output for protection-related messages (e.g., "protected branch", "required status check", "required review") and surface a specific warning with guidance to configure `checkpoint_remote` to a different remote.

2. **Proactive** (optional, via `gh api`): Before pushing, check if the target ref matches any branch protection rules on the remote. This avoids the failed push attempt entirely but adds an API dependency.

The reactive approach is simpler and aligns with Partio's minimal-dependency philosophy.

## Context

- Inspired by entireio/cli PR #1033 which addresses the same problem
- Relevant code: `internal/hooks/prepush.go` — the `runPrePush` function
- Related: `internal/git/` for push operations
- The `checkpoint_remote` config option (not yet implemented in Partio) would be the recommended workaround

## Why This Matters

Users in organizations with strict branch protection encounter confusing errors or silent failures when Partio tries to push checkpoint data. This erodes trust in the tool and makes it harder to adopt in enterprise environments where branch protection is standard.

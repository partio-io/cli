---
id: preserve-checkpoints-across-ff-merges
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint references remain discoverable after fast-forward merges to the target branch"
  - "The checkpoint branch DAG correctly links checkpoints to their commits regardless of which branch name points at them"
  - "`partio status` and checkpoint queries find sessions for commits that arrived via fast-forward merge"
  - "No duplicate checkpoints are created during the merge process"
pr_labels:
  - minion
  - minion-proposal
---

# Preserve checkpoint references across fast-forward merges

## Description

Ensure that checkpoint references remain fully discoverable after fast-forward merges. Currently, when a feature branch with Partio-Checkpoint trailers is fast-forward merged into the main branch, the checkpoint metadata on the orphan branch still points at the correct commit SHAs (since FF merges preserve commit identity). However, session lookup and checkpoint queries may not correctly resolve these references when the branch context changes.

Add integration tests that verify the full checkpoint discovery path after FF merges, and fix any lookup logic that assumes checkpoints are only associated with the branch where they were originally created. This complements the existing squash-merge handling (proposal #338) to cover all common merge strategies.

## Why

Fast-forward merges are the default merge strategy for many teams (especially those using rebase workflows). If checkpoint references silently break after merging, users lose the ability to trace AI-authored code back to its session context on their main branch — which is exactly where they need it most for code review and audit.

## Context hints

- `internal/checkpoint/` — checkpoint storage and query logic
- `internal/hooks/` — post-commit and pre-push hook implementations
- `internal/git/` — git operations and branch management

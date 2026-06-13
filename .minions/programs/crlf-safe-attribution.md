---
id: crlf-safe-attribution
target_repos:
  - cli
acceptance_criteria:
  - "Attribution diff comparison respects git's clean/smudge filters and `core.autocrlf` setting"
  - "Files with only line-ending differences (CRLF vs LF) are not counted as agent-modified"
  - "Checkpoint content comparison uses `git diff --quiet` as the primary cleanliness check before falling back to raw blob hashing"
  - "Regression test verifies that `core.autocrlf=true` does not inflate attribution percentages"
  - "Windows and cross-platform CI environments produce consistent attribution results"
pr_labels:
  - minion
---

# Handle line-ending normalization in attribution and checkpoint comparison

## Description

Update file comparison logic in attribution calculation and checkpoint content detection to respect git's line-ending normalization settings (`core.autocrlf`, clean/smudge filters). Currently, raw byte comparison between working tree files and committed blobs can produce false positives when git normalizes line endings on commit.

The fix should:
1. Use `git diff --quiet <path>` as the primary check for whether a file has meaningful changes, before falling back to raw blob hash comparison
2. Ensure attribution calculations don't inflate agent percentages due to CRLF/LF differences
3. Add a regression test that enables `core.autocrlf=true` and verifies line-ending-only changes don't affect attribution

## Why

On Windows or in cross-platform teams where `core.autocrlf=true` is common, git normalizes line endings on commit (CRLF in working tree → LF in repository). If Partio compares raw on-disk bytes to committed blob hashes, every file touched by the agent could appear "modified" even when the only difference is line endings. This leads to inflated attribution percentages and potentially spurious checkpoint content.

## Source

entireio/cli PR #913 — Fix false carry-forward detection under `core.autocrlf=true`

## Context hints

- `internal/attribution/` — attribution calculation logic
- `internal/checkpoint/` — checkpoint content comparison
- `internal/git/` — git operations wrapper

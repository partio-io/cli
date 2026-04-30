---
id: handle-line-ending-normalization
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint diff operations produce correct results when core.autocrlf is enabled
  - Transcript content stored in checkpoints uses normalized (LF) line endings regardless of platform
  - Attribution calculation is not affected by CRLF vs LF differences in working tree files
  - Tests cover scenarios with autocrlf=true on both Windows and Unix-like systems
pr_labels:
  - minion
---

# Handle Git line-ending normalization in checkpoint operations

## Problem

When Git's `core.autocrlf` or `.gitattributes` line-ending normalization is active, the working tree may contain CRLF-encoded files while committed blobs use LF. Partio's checkpoint operations (diff calculation, attribution, transcript storage) may compare raw on-disk bytes against committed blob hashes, producing false differences or incorrect attribution when line endings differ only due to normalization.

## Desired Behavior

1. When comparing working tree content to committed content for checkpoint purposes, ask Git whether the path is clean (via `git diff --name-only` or `git status --porcelain`) rather than doing raw byte comparison.
2. Normalize transcript content to LF before storing in checkpoint blobs, ensuring consistent storage regardless of the platform where the checkpoint was created.
3. Ensure attribution line-count calculations are not inflated by CRLF/LF differences.

## Context Hints

- `internal/checkpoint/` — checkpoint creation and diff storage
- `internal/attribution/` — line-count attribution calculation
- `internal/git/` — git operations wrapper

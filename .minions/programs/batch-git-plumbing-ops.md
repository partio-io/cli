---
id: batch-git-plumbing-ops
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint creation uses batched git hash-object calls (--stdin-paths or similar) instead of one subprocess per blob
  - mktree accepts pre-computed object hashes without re-hashing
  - Benchmark shows measurable reduction in subprocess fork count for a typical checkpoint write
  - Existing checkpoint format and content remain unchanged
pr_labels:
  - minion
  - performance
---

# Batch git plumbing subprocess calls during checkpoint creation

## Problem

Each checkpoint write currently spawns separate git subprocesses for hash-object, mktree, commit-tree, and update-ref. On repos with large sessions or frequent commits, the per-fork overhead accumulates — especially on platforms where process creation is expensive (Windows, resource-constrained CI).

## Proposed Solution

Batch git plumbing operations where the git CLI supports it:

1. **hash-object**: Use `git hash-object --stdin-paths` to hash multiple blobs in a single invocation, or pipe content via stdin with `--stdin`.
2. **mktree**: Feed all tree entries in a single invocation rather than building incrementally.
3. **Connection reuse**: Where possible, reuse a single git process for multiple related operations (e.g., `git cat-file --batch` pattern for reads).

The orphan branch structure and checkpoint format must remain identical — this is purely a performance optimization of the write path.

## Context

- `internal/checkpoint/` — checkpoint storage implementation using git plumbing
- `internal/git/` — git operation wrappers

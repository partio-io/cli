---
id: smart-checkpoint-fetch-strategy
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint fetches detect whether the repo is a partial clone (via extensions.partialClone config)
  - Partial-clone repos use --filter=blob:none for checkpoint fetches
  - Non-partial-clone repos use --depth=1 (shallow) for checkpoint fetches
  - Repos that need full history for rebase/merge use plain fetches
  - No repo is unintentionally converted to a promisor repo by checkpoint operations
pr_labels:
  - minion
---

# Adaptive fetch strategy for checkpoint refs based on repo type

## Summary

Checkpoint fetches currently may use `--filter=blob:none` unconditionally, which converts non-promisor repositories into promisor-enabled ones as a side effect. This is an unexpected and potentially disruptive change to the repository's git configuration.

## What to implement

Before fetching checkpoint refs, detect whether the repository is already a partial clone by checking the `extensions.partialClone` git config key:
- If the repo is already a partial clone: use `--filter=blob:none` (treeless fetch) for consistency
- If the repo is not a partial clone: use `--depth=1` (shallow fetch) to minimize data transfer without changing the repo type
- For operations that need full history (e.g., rebase-based checkpoint alignment): use a plain fetch

## Why this matters

Converting a repo to a promisor repo has side effects that affect all git operations, not just Partio's. Some CI systems and tools don't support promisor repos well. Users should not have their repo type changed as a side effect of installing a checkpoint tool.

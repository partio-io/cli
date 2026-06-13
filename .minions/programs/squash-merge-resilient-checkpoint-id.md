---
id: squash-merge-resilient-checkpoint-id
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint ID can optionally be embedded in the commit title as a short suffix (e.g., [ckpt:a3f9b2c1])
  - A configuration option controls whether title embedding is enabled (default off)
  - When enabled, prepare-commit-msg writes both the trailer and the title suffix
  - Checkpoint lookup can resolve commits by either trailer or title suffix
  - Squash merges that preserve only the commit title retain the checkpoint link
pr_labels:
  - minion
---

# Store checkpoint ID in commit title for squash merge resilience

## Summary

Platforms like GitLab discard the commit message body during squash merges, preserving only the commit title (first line). This strips the `Partio-Checkpoint:` trailer and permanently breaks the link between the squashed commit and its checkpoint metadata.

## What to implement

Add a configurable option (`checkpoint_title_format: true` in settings) that embeds a short checkpoint ID in the commit title as a suffix: `feat: add login [ckpt:a3f9b2c1]`. When enabled, the prepare-commit-msg hook writes both the standard trailer and the title suffix.

Checkpoint resolution logic should be updated to search for both trailer-based and title-based checkpoint IDs, falling back to title parsing when no trailer is present.

## Why this matters

Teams using squash merge workflows (common on GitLab, also used on GitHub) lose all checkpoint linkage when merging to main. This makes the checkpoint data unreachable for exactly the commits that matter most — the ones on the main branch. A title-embedded ID survives any merge strategy since all platforms preserve the first line.

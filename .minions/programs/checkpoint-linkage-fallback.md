---
id: checkpoint-linkage-fallback
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint metadata includes tree_hash and patch_id fields computed from the linked commit
  - When a commit's Partio-Checkpoint trailer is stripped (squash merge, rebase, reword), the checkpoint can still be matched by tree hash or patch ID
  - Tree hash is the git tree object hash of the commit being checkpointed
  - Patch ID is computed via git patch-id from the commit's diff
  - Existing checkpoints without fallback fields continue to work (backwards compatible)
  - Unit tests verify linkage metadata is written and can be used for matching
pr_labels:
  - minion
---

# Add tree hash and patch ID as checkpoint linkage fallback

Store git-native linkage signals (tree hash and patch ID) in checkpoint metadata so that checkpoints can still be matched to commits even when the `Partio-Checkpoint` trailer is stripped by history rewrites, squash merges, or rebases.

## Context

Partio links commits to checkpoints via the `Partio-Checkpoint` trailer in the commit message. Multiple scenarios strip this trailer: GitLab squash merges discard the commit body, interactive rebases with `reword` can drop trailers, and `git filter-branch` rewrites may lose them.

When the trailer is gone, the checkpoint becomes orphaned — the session context is preserved but unreachable from the commit.

## Approach

During post-commit checkpoint creation, compute and store two additional fields in the checkpoint metadata:

1. **`tree_hash`** — the git tree object hash of the committed state (`git rev-parse HEAD^{tree}`)
2. **`patch_id`** — the content-based diff fingerprint (`git diff-tree HEAD | git patch-id`)

These are git-native identifiers that survive history rewrites: a squash merge produces the same tree hash if the final content matches, and a rebased commit produces the same patch ID if the diff content is unchanged.

Store these in the checkpoint's metadata JSON alongside the existing `commit_hash` field. Downstream consumers (UI, `partio rewind`) can use them as fallback matching signals when the trailer lookup fails.

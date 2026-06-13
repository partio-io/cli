---
id: track-checkpoint-attach-reason
target_repos:
  - cli
pr_labels:
  - minion
acceptance_criteria:
  - Checkpoint metadata includes an "attach_reason" field indicating why the checkpoint was created
  - The post-commit hook sets the attach reason to "manual-commit" for the manual-commit strategy
  - The attach reason is included in partio checkpoint list output (text and JSON modes)
  - The metadata schema is documented and the field is preserved through checkpoint storage round-trips
  - Unit tests verify attach_reason is set correctly for the manual-commit strategy
---

# Track attach reason in checkpoint metadata

Record why each checkpoint was created (manual-commit, auto-commit, condensation, import, etc.) in the checkpoint metadata, providing auditability for how checkpoints were triggered.

## Context

Entireio/cli PR #1199 adds an attach reason to checkpoint metadata, allowing PostCommit to record per-session whether the checkpoint was a normal commit, a condensation-triggered save, or another trigger. As Partio adds more capture strategies and features like bulk import, knowing *why* a checkpoint was created becomes important for filtering, debugging, and display.

## What to implement

1. Add an `AttachReason` field to the `Checkpoint` domain type in `internal/checkpoint/`.
2. Set the reason in the post-commit hook based on the active strategy and trigger (e.g., "manual-commit").
3. Persist the reason in checkpoint metadata when writing to the orphan branch.
4. Include the attach reason in `partio checkpoint list` output.
5. Define reason constants for current and anticipated triggers: `manual-commit`, `import`, `rewind`.

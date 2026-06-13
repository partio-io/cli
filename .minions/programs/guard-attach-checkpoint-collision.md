---
id: guard-attach-checkpoint-collision
target_repos:
  - cli
acceptance_criteria:
  - partio attach detects when a checkpoint ID already exists on the checkpoint branch and refuses to overwrite it
  - partio attach generates a new unique checkpoint ID when the trailer-referenced ID is already taken
  - a warning message is displayed explaining the collision and the new ID used
  - existing attach behavior is preserved when no collision exists
pr_labels:
  - minion
---

# Guard `partio attach` against overwriting checkpoints from other machines

## Problem

When a user commits with Partio on machine A and pushes the code commit (with `Partio-Checkpoint` trailer) but does not push the checkpoint branch, a second user on machine B who pulls the commit and runs `partio attach` will reuse the checkpoint ID from the trailer. This silently overwrites machine A's checkpoint data when both eventually push their checkpoint branches, causing data loss.

## Desired behavior

`partio attach` should check whether the checkpoint ID referenced in the commit trailer already exists on the checkpoint branch (locally or via a fetch). If it does, attach should:

1. Warn the user that a checkpoint with that ID already exists from another source
2. Generate a new unique checkpoint ID for the local attach operation
3. Update the commit trailer to reference the new ID
4. Proceed with creating the checkpoint under the new ID

This prevents silent data clobbering while still allowing the attach workflow to succeed.

## Context

- Inspired by entireio/cli PR #1014 which addresses the same collision scenario
- The `partio attach` command is proposed in partio-io/cli#158
- Checkpoint IDs are currently derived from commit hashes; collision happens when the same commit exists on multiple machines with different session data
- The checkpoint branch (`partio/checkpoints/v1`) uses git plumbing writes, so collision detection requires checking existing refs/trees

## Implementation hints

- In `internal/checkpoint/`, add a check before writing: verify the checkpoint ref doesn't already exist
- Use `git rev-parse` or equivalent to check if the ref `partio/checkpoints/v1:<checkpoint-id>` resolves
- Consider fetching the remote checkpoint branch first if `push_sessions` is enabled
- Generate new IDs using the existing UUID or hash-based approach

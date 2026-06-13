---
id: filtered-checkpoint-ref-fetches
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint fetch operations use refspec filtering to only download refs under partio/checkpoints/v1
  - Clone and fetch of checkpoint data transfers significantly less data than a full ref fetch
  - Existing checkpoint push and read operations continue to work correctly with filtered refs
  - partio doctor validates that the filtered refspec is correctly configured
pr_labels:
  - minion
---

# Use refspec filtering for checkpoint ref fetches

## Problem

When Partio fetches or clones checkpoint data from a remote, it currently transfers all refs, including those unrelated to checkpoints. As repositories accumulate history, this becomes increasingly expensive in terms of network transfer and time — especially for large teams where the checkpoint branch grows significantly.

## Solution

Add refspec filtering to all checkpoint fetch and clone operations so that only refs under `partio/checkpoints/v1` (and any future checkpoint ref patterns) are transferred. This reduces the data transferred during `git push` (which triggers checkpoint sync) and any explicit checkpoint fetch operations.

### Implementation hints

- In `internal/checkpoint/` and `internal/git/`, update fetch/clone calls to use filtered refspecs (e.g., `+refs/partio/checkpoints/*:refs/remotes/origin/partio/checkpoints/*`)
- Ensure `pre-push` hook checkpoint sync uses the filtered refspec
- Add a doctor check to validate that the checkpoint remote has the correct refspec configured
- Consider making this opt-in initially via a config setting if backward compatibility is a concern

## Inspiration

Adapted from entireio/cli 0.5.6 (#996) which added filtered fetches for checkpoint refs to reduce clone/fetch size.

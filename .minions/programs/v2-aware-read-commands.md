---
id: v2-aware-read-commands
target_repos:
  - cli
acceptance_criteria:
  - "`partio rewind` resolves checkpoints from v2 store first, falling back to v1 when v2 data is not available"
  - "`partio status` displays checkpoint information from whichever store (v1 or v2) contains the data"
  - "Listing checkpoints merges results from both v1 and v2 stores so pre-migration checkpoints remain visible"
  - "When v2 compact transcript exists, read commands use it instead of the full JSONL"
  - "A test verifies that a v1-only checkpoint is still accessible after v2 support is added"
  - "A test verifies that a v2 checkpoint is preferred over a v1 checkpoint when both exist"
pr_labels:
  - minion
  - feature
---

# Make read commands v2-aware with v1 fallback

Once checkpoints v2 (compact `transcript.jsonl`) is implemented for writes, all read-path commands (`partio rewind`, `partio status`, and any future checkpoint inspection commands) must be updated to resolve checkpoints from the v2 store first, falling back to v1 for older checkpoints.

## Why

Without this, commands that read checkpoint data will break or return empty results once checkpoints are primarily written in v2 format. Users who upgrade will lose visibility into new checkpoints through the CLI, even though the data exists. Additionally, pre-migration v1 checkpoints must remain accessible during the transition period.

## Approach

- Add a resolution layer in the checkpoint read path that checks v2 first, then v1
- When listing checkpoints, merge results from both stores (deduplicating by checkpoint ID)
- When reading transcript data, prefer the compact v2 `transcript.jsonl` over the full session JSONL
- Ensure `partio rewind` works with both checkpoint formats

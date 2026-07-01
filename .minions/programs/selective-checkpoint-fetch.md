---
id: selective-checkpoint-fetch
target_repos:
  - cli
acceptance_criteria:
  - "Pre-push hook supports fetching only new checkpoint refs instead of the full branch"
  - "Checkpoint fetch uses depth-limited or ref-specific git fetch when possible"
  - "Falls back to full fetch when selective fetch fails"
  - "Reduces network transfer for repos with large checkpoint histories"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add selective checkpoint fetch for efficient push sync

Currently, checkpoint push/fetch operations transfer the entire `partio/checkpoints/v1` branch history. For repositories with many checkpoints, this becomes increasingly slow.

## Desired behavior

- When pushing checkpoints in the pre-push hook, fetch only the remote branch tip (not the full history) to determine if a fast-forward is possible
- Use `--depth=1` or ref-specific fetch to minimize network transfer
- Fall back to full fetch if the shallow approach fails (e.g., force-push scenarios)

## Implementation notes

- The checkpoint push logic lives in `internal/hooks/` (pre-push implementation)
- Checkpoint storage uses git plumbing in `internal/checkpoint/`
- The key optimization is in the fetch step before push — currently it may fetch more history than needed
- Add a config option `strategy_options.shallow_checkpoint_fetch` (default: true) to allow disabling
- Preserve the existing error resilience pattern: log warnings but don't block git operations

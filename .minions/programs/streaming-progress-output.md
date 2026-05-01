---
id: streaming-progress-output
target_repos:
  - cli
acceptance_criteria:
  - Long-running operations (rewind, checkpoint push) display incremental progress to the terminal
  - Progress updates use carriage-return overwriting on TTY, plain newlines otherwise
  - Non-TTY output (piped, CI) degrades gracefully to simple status lines
  - No visible change for operations that complete in under 1 second
pr_labels:
  - minion
  - enhancement
---

# Add streaming progress output for long-running operations

Operations like `partio rewind` (which fetches and checks out checkpoint data) and checkpoint push during pre-push can take several seconds. Currently they produce no output until completion, leaving users uncertain whether the CLI is stuck.

## What to implement

1. Add a progress writer utility that:
   - Detects TTY vs pipe
   - On TTY: uses `\r` to overwrite the current line with status updates
   - On non-TTY: emits one-line status messages at key milestones
2. Integrate progress reporting into `partio rewind` and the pre-push checkpoint sync
3. Suppress progress output for operations completing in < 1 second to avoid flicker

## Why this matters

Users need feedback that the CLI is working, especially during network operations. Silent delays erode trust and prompt users to Ctrl+C or disable features.

## Source

Inspired by entireio/cli PR #964 — stream-driven progress for explain --generate.

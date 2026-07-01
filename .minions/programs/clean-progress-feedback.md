---
id: clean-progress-feedback
target_repos:
  - cli
acceptance_criteria:
  - partio clean and partio reset show progress output when processing checkpoint data
  - Progress includes count of items processed and total items when known
  - Operations can be interrupted with Ctrl+C without leaving the checkpoint branch in a corrupted state
  - Non-TTY environments (CI, pipes) receive simplified progress without terminal control codes
  - Operations complete without hanging on large checkpoint histories
pr_labels:
  - minion
---

# Add progress feedback and safe interruption to partio clean and reset

## Summary

Add progress output to `partio clean` and `partio reset` commands so users get feedback during long-running operations on large checkpoint histories, and ensure operations can be safely interrupted without corrupting the checkpoint branch.

## Motivation

When a repository accumulates many checkpoints over weeks or months, `partio clean` and `partio reset` can take significant time as they walk and modify the checkpoint branch. Currently these commands produce no output during processing, making them appear hung. Users may kill the process, potentially leaving the checkpoint branch in an inconsistent state.

Inspired by entireio/cli#1182 which added progress and an escape hatch to `entire clean --all` after it was observed to hang silently for many seconds on large histories.

## Design

1. **Progress output**: During checkpoint branch traversal, print periodic progress lines to stderr:
   ```
   Cleaning checkpoints... 45/128
   Cleaning checkpoints... 90/128
   Done. Removed 128 checkpoints.
   ```

2. **TTY detection**: Use terminal width detection to show an updating progress line on TTYs, or simple periodic lines on non-TTY outputs.

3. **Safe interruption**: Handle SIGINT (Ctrl+C) gracefully — complete the current git object operation, then stop and report how many items were processed. The checkpoint branch should remain valid after interruption.

4. **Timeout guard**: Add a configurable timeout (default: 60s) to prevent indefinite hangs when git operations stall. Report the timeout clearly if hit.

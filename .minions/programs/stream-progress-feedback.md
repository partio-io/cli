---
id: stream-progress-feedback
target_repos:
  - cli
acceptance_criteria:
  - Long-running CLI operations show real-time progress instead of silent waiting
  - Progress output is TTY-gated — only shown when stderr is a terminal
  - Non-TTY environments (CI, pipes) get a single start/complete line instead of progress updates
  - At minimum, checkpoint creation and push operations show progress indicators
  - Progress messages are written to stderr so they don't interfere with stdout data
  - No external dependencies added for progress display
pr_labels:
  - minion
---

# Add real-time progress feedback for long-running CLI operations

Show real-time progress indicators for checkpoint creation and push operations instead of silent waiting, with TTY-gated output that degrades gracefully in non-interactive environments.

## Context

Several Partio operations can take noticeable time: creating checkpoints (especially with large session transcripts), pushing checkpoint branches to remotes, and any future operations like rewind or search. Currently these operations run silently, leaving users uncertain whether the tool is working or hung.

## Approach

Add a lightweight progress feedback mechanism:

1. **TTY detection**: Check if stderr is a terminal (`os.Stderr.Fd()` + `term.IsTerminal()` or equivalent)
2. **Interactive mode**: When stderr is a TTY, show inline progress updates (e.g., `Creating checkpoint... (transcript: 47.0 KB)`, `Pushing checkpoint branch...`) using carriage returns for in-place updates
3. **Non-interactive mode**: When stderr is not a TTY, emit a single start line and a completion line — no progress animation

Apply this to the most visible operations first:
- Post-commit checkpoint creation (especially the git plumbing write sequence)
- Pre-push checkpoint branch push

Use Go's standard library only — no external TUI frameworks. The progress output should be minimal and informative, not decorative.

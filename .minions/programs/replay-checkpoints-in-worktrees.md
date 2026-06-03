---
id: replay-checkpoints-in-worktrees
target_repos:
  - cli
acceptance_criteria:
  - partio replay <ref> creates an isolated worktree at the checkpoint's parent commit
  - Prompt sequence is correctly extracted from the checkpoint's JSONL transcript
  - --dry-run outputs the extracted prompts without executing anything
  - Replay worktrees are cleaned up after use or on --cleanup
  - Command fails gracefully with a clear message if the checkpoint ref doesn't exist
  - Unit tests cover prompt extraction and worktree lifecycle
pr_labels:
  - minion
---

# Add partio replay command for checkpoint playback in isolated worktrees

Add a `partio replay` command that replays checkpoint sessions in isolated git worktrees, enabling users to evaluate, compare, and debug AI agent sessions captured by Partio.

## Implementation Notes

- `partio replay <checkpoint-ref>` creates an isolated git worktree at the checkpoint's base commit
- Extracts the prompt sequence from the checkpoint's JSONL transcript
- Optionally feeds prompts to a running agent session (or outputs them for manual replay)
- Saves a report comparing the replay output against the original checkpoint's file changes
- Cleans up the worktree after completion
- Flags: `--report`, `--dry-run`, `--cleanup`

## Context

- `internal/checkpoint/` — checkpoint retrieval and metadata
- `internal/agent/claude/parse_jsonl.go` — extracting prompts from transcripts
- `internal/git/` — worktree creation and management
- `cmd/partio/` — CLI command registration

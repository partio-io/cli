---
id: normalize-machine-paths-in-transcripts
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint transcripts replace absolute repo root paths with a placeholder (e.g., `$REPO_ROOT`) before storage"
  - "Checkpoint transcripts replace home directory paths with a placeholder (e.g., `$HOME`) before storage"
  - "Users can configure additional path replacement rules via `transcript_filters` in settings"
  - "When displaying transcript content (e.g., in rewind or future browsing commands), placeholders are expanded back to the current machine's paths"
  - "Existing checkpoints without path normalization continue to display correctly"
  - "Path normalization does not alter non-path content in transcripts"
pr_labels:
  - minion
---

# Normalize machine-specific paths in checkpoint transcripts

## Summary

Add a clean/smudge pipeline that normalizes machine-specific absolute paths (repo root, home directory) in session transcripts before they are stored in checkpoints, and restores them when displayed. This makes checkpoint data portable across machines and reduces noise from environment-specific paths.

## Motivation

Session transcripts captured by partio contain absolute file paths that are specific to the machine where the agent ran (e.g., `/home/alice/projects/myapp/src/main.go`). These paths create problems:

1. **Portability**: Checkpoints viewed on a different machine or by a different user show paths that don't exist locally
2. **Privacy**: Home directory paths leak usernames and directory structure
3. **Noise**: Long absolute paths obscure the relevant relative path information
4. **Diffability**: The same logical session on two machines produces different checkpoint content

Inspired by entireio/cli PR #758 which implements a git-like clean/smudge pipeline for this purpose.

## Design Notes

- Implement as a `filter` or `normalize` package with `Clean(content, config)` and `Smudge(content, config)` functions
- Default replacements: repo root -> `$REPO_ROOT`, home dir -> `$HOME`
- Apply `Clean` in the checkpoint write path (before content is hashed and stored)
- Apply `Smudge` in any display/read path
- Support user-configured additional filters in settings: `transcript_filters: [{match: "/custom/path", replace: "$CUSTOM"}]`
- Must be backwards-compatible: transcripts without placeholders should display as-is

## Context Hints

- `internal/checkpoint/` — checkpoint write and read paths
- `internal/agent/claude/parse_jsonl.go` — transcript parsing
- `internal/config/` — settings for configurable filters

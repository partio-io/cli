---
id: committed-checkpoint-metadata
target_repos:
  - cli
acceptance_criteria:
  - New config option (e.g., committed_metadata=true) enables writing a lightweight metadata file on commit
  - Metadata file (e.g., .partio/checkpoint.json) contains session ID, attribution percentage, and prompt summary
  - File is automatically staged and included in the commit (or amend) so it appears in diffs and PRs
  - Orphan branch checkpoint storage continues to work unchanged (this is additive)
  - File content updates on each commit, not accumulates (single file, not one per commit)
pr_labels:
  - minion
  - feature
---

# Optional committed checkpoint metadata alongside code

## Problem

Checkpoint data stored exclusively on the orphan branch is invisible during normal git operations — it doesn't show up in diffs, PRs, or code review. Reviewers and teammates cannot see at a glance whether a commit was AI-assisted, what the session context was, or what attribution looks like without running `partio` commands.

## Proposed Solution

Add an optional mode where a lightweight metadata file is written as a committed file in the working tree:

- File location: `.partio/checkpoint.json` (or configurable)
- Contains: session ID, agent name, attribution percentage, prompt summary (first line of the user's request)
- Updated on each post-commit hook run (overwrites previous content)
- Automatically staged as part of the commit amend that adds trailers

This makes checkpoint context visible in:
- `git diff` and `git log -p`
- Pull request diffs on GitHub/GitLab
- Code review tools

The orphan branch remains the authoritative store for full session data. This file is a lightweight signal for visibility.

## Context

- `internal/hooks/` — post-commit hook that creates checkpoints
- `internal/checkpoint/` — checkpoint domain and storage
- `internal/config/` — where the new option would be registered

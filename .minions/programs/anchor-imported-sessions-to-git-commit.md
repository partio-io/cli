---
id: anchor-imported-sessions-to-git-commit
target_repos:
  - cli
pr_labels:
  - minion
acceptance_criteria:
  - "`partio import` stamps each imported checkpoint with a `commit_sha` field resolving in order: origin/HEAD → local default branch HEAD → current HEAD → empty string"
  - "The `commit_sha` is stored in the checkpoint metadata alongside the existing session fields"
  - "`partio rewind` displays imported checkpoints in the correct position relative to the associated commit"
  - "If no commit SHA can be resolved, import succeeds with an empty `commit_sha` field rather than failing"
  - "Tests cover the SHA resolution priority order using `t.TempDir()` git repos"
---

# Anchor imported checkpoints to the git commit at import time

## Description

When `partio import` brings in pre-existing Claude Code sessions as checkpoints, those checkpoints have no connection to the git commit timeline — they float free in the orphan branch with no anchor. This makes them invisible or confusingly positioned when browsing with `partio rewind`.

Partio should stamp each imported checkpoint with the SHA of the relevant commit at import time: prefer `origin/HEAD`, fall back to the local default branch head, then `HEAD`, and leave empty if nothing resolves. Store this SHA in a `commit_sha` field in the checkpoint metadata (alongside the existing `session_id`, `attribution`, etc. fields).

This mirrors the approach taken in entireio/cli#1825 and gives `partio rewind` a join key to show imported sessions in the right historical context.

### Desired behaviour

```
$ partio import
Scanning for Claude Code sessions...
Found 3 session(s) to import.
Anchoring to commit abc1234 (origin/main HEAD)
Imported 3 checkpoint(s).
```

### Implementation notes

- The SHA resolution lives in `cmd/partio/` (the import command), not in `internal/checkpoint/` — keep domain types clean.
- `internal/checkpoint/` stores the field; the write plumbing already uses `git hash-object` / `commit-tree` — add `commit_sha` to the JSON blob.
- Read the SHA via `git rev-parse origin/HEAD` (shell exec, same pattern as other git operations in the codebase), falling through the priority chain on error.

<!-- program: .minions/programs/anchor-imported-sessions-to-git-commit.md -->

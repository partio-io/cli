---
id: partio-search-checkpoints
target_repos:
  - cli
acceptance_criteria:
  - "`partio search <query>` returns checkpoints whose metadata (prompt, context, files touched, commit message) contains the query string"
  - "Results show checkpoint ID, associated commit hash (short), commit date, and a matching excerpt"
  - "`--since <date>` flag filters results to checkpoints created after the given date (ISO 8601)"
  - "Returns exit code 1 with a clear message when no checkpoints are found or the repo has no checkpoint branch"
  - "Matches are case-insensitive"
pr_labels:
  - minion
---

# Add `partio search` command to query checkpoint history

## What

Add a `partio search <query>` command that searches across checkpoint metadata stored on the `partio/checkpoints/v1` branch. The command reads checkpoint trees from the orphan branch and matches against available fields: prompt, context, files touched, and associated commit messages.

Example usage:
```
partio search "authentication"
partio search "refactor database" --since 2026-01-01
```

Output example:
```
checkpoint abc1234  →  commit def5678  (2026-03-15)
  Files: internal/auth/handler.go, internal/auth/middleware.go
  Prompt: Refactor the authentication middleware to use...
```

## Why

As checkpoint data accumulates across many sessions, users need a way to find relevant historical AI sessions without scrolling through `git log`. Search enables knowledge retrieval: "when did we refactor the auth layer?" or "what sessions touched the checkpoint storage code?" This is the core value proposition of preserving AI session context — it should be queryable.

## Source

Inspired by `entireio/cli` PR #833 which scaffolds managed search subagents for querying checkpoint data.

## Implementation hints

- `cmd/partio/search.go` — new Cobra subcommand
- `internal/checkpoint/` — for reading checkpoint trees from the orphan branch
- Use existing git plumbing (ls-tree, cat-file) to iterate checkpoints without checkout
- Start with simple substring matching; avoid external search dependencies (keep minimal deps)
- Checkpoint metadata is stored in JSON blobs within the checkpoint trees
<!-- program: .minions/programs/partio-search-checkpoints.md -->

---
id: author-scoped-activity-reports
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint queries can be filtered by commit author email"
  - "A --me flag resolves to the current user's git config user.email"
  - "--me and --author are mutually exclusive with a clear error message"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add author-scoped filtering for checkpoint queries

Add `--me` and `--author <email>` flags to checkpoint listing and status commands so users can filter checkpoint activity to their own work or a specific contributor's work.

## What to implement

1. Add `--me` flag that resolves to the current user's `git config user.email` at runtime.
2. Add `--author <email>` flag for filtering by a specific contributor's email.
3. Make `--me` and `--author` mutually exclusive with a clear error message.
4. Apply author filtering by matching against the commit author of the linked git commits.
5. Add the flags to relevant commands that list or display checkpoint data (e.g., checkpoint list, status).

## Context hints

- `internal/checkpoint/` — checkpoint domain and storage
- `cmd/partio/` — CLI command definitions
- `internal/git/` — git operations (for resolving user.email)

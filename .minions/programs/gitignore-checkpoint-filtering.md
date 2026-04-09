---
id: gitignore-checkpoint-filtering
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint creation filters file lists through git check-ignore before writing tree objects"
  - "Gitignored files (e.g. .env, node_modules/, *.secret) never appear in checkpoint tree objects"
  - "First checkpoint (based on git status) remains unaffected since git status already respects .gitignore"
  - "Table-driven tests verify gitignored paths are excluded from checkpoint data"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Filter gitignored files from checkpoint data

## Problem

Checkpoint tree objects may include files that should be gitignored. The first checkpoint in a session is safe because it uses `git status` which respects `.gitignore`. However, for subsequent checkpoints, file lists come from agent transcript extraction and are passed through to tree building without gitignore filtering. If an agent touches a gitignored file like `.env`, it could appear in checkpoint data.

Checkpoint branches are local and temporary, but there is a risk they get pushed to a remote (e.g., via `git push --all`), and gitignored files should never be persisted in checkpoints regardless.

## Solution

Add a `git check-ignore` filter step in the checkpoint storage layer. Before writing tree objects, filter the file list (ModifiedFiles/NewFiles from session data) through `git check-ignore` and exclude any matched paths.

## Context

- Inspired by entireio/cli#890
- Relevant code: `internal/checkpoint/` (storage layer), `internal/session/` (session data with file lists)

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```

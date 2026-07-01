---
id: async-post-commit-checkpoint
target_repos:
  - cli
acceptance_criteria:
  - Post-commit hook spawns a detached background process for checkpoint creation and returns immediately
  - Git commit operations complete without waiting for transcript parsing or blob storage
  - Background process logs errors to a file instead of blocking the terminal
  - A lock file prevents concurrent background checkpoint writers from corrupting refs
  - partio status shows pending/in-progress checkpoint writes
  - If the background process fails, the next post-commit hook retries the missed checkpoint
pr_labels:
  - minion
---

# Defer post-commit checkpoint creation to a background process

## Summary

Move the heavy work in the post-commit hook (transcript JSONL parsing, git blob/tree/commit creation, ref updates, and commit amend for trailers) to a detached background process. The hook itself should only save minimal state and spawn the worker, returning control to the user immediately.

## Why

Multiple reports in entireio/cli (#957, #1072, #1137) show that synchronous hook execution causes 30-second timeouts and noticeable latency during agent coding sessions. Partio's post-commit hook currently performs transcript parsing, attribution calculation, git plumbing operations, and commit amendment all synchronously — any slowness in these steps directly delays the developer's git workflow.

## What to implement

1. **Split post-commit into dispatch + worker**: The post-commit hook saves detection state (as it does now) and spawns a detached `partio _checkpoint-worker` process, then exits immediately.

2. **Background worker**: A new hidden subcommand (`partio _checkpoint-worker`) reads the saved state, performs transcript parsing, creates the checkpoint on the orphan branch, and amends the commit with trailers.

3. **Lock file**: Use a lock file in `.partio/state/` to prevent concurrent workers from corrupting checkpoint refs when rapid commits occur.

4. **Retry on failure**: If the worker fails, persist the pending state so the next post-commit invocation can retry.

5. **Status visibility**: `partio status` should show if there are pending checkpoint writes that haven't completed yet.

## Context hints

- `internal/hooks/post_commit.go` — current synchronous implementation
- `internal/checkpoint/` — checkpoint creation logic
- `internal/session/` — session state persistence
- `cmd/partio/` — CLI command registration

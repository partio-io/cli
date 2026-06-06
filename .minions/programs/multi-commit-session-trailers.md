---
id: multi-commit-session-trailers
target_repos:
  - cli
pr_labels:
  - minion
  - bug
acceptance_criteria:
  - When an agent session spans multiple manual commits, each commit receives a Partio-Checkpoint trailer
  - The post-commit hook correctly creates separate checkpoints for each commit within the same session
  - Session state is preserved across commits within the same active session (not deleted after the first commit)
  - Unit tests verify that two sequential commits during the same active session both receive trailers and checkpoints
---

# Ensure all commits in a multi-commit session receive checkpoint trailers

Fix the hook state lifecycle so that every commit made during an active agent session gets a `Partio-Checkpoint` trailer, not just the first one.

## Context

Entireio/cli issue #784 reports that only the first commit per session gets the `Entire-Checkpoint` trailer; subsequent commits are silently skipped. Partio likely has the same issue: the post-commit hook reads and immediately deletes the pre-commit state file (`.partio/state/pre-commit.json`), so the second commit in the same session finds no state and skips checkpoint creation.

## What to implement

1. Modify the post-commit hook to re-detect the active agent session directly (via the detector interface) rather than relying solely on the pre-commit state file, OR preserve the state file across commits while the session remains active.
2. Ensure each commit within a session gets its own checkpoint on the orphan branch, with the correct commit SHA in metadata.
3. Keep the existing re-entry prevention (delete state before amend) intact — the fix must not break the `git commit --amend` re-trigger guard.

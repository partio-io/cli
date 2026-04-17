---
id: fetch-checkpoint-refs-by-url
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint ref fetches use the remote URL directly instead of the remote name
  - No new refspec entries are added to .git/config when fetching checkpoints
  - Existing checkpoint fetch functionality (pre-push, rewind) continues to work correctly
  - Integration test verifies .git/config is not modified by checkpoint fetch operations
pr_labels:
  - minion
---

# Fetch checkpoint refs by URL to avoid polluting git config

## Summary

When Partio fetches checkpoint refs (e.g., `partio/checkpoints/v1`), using the remote name causes git to add persistent refspec entries to `.git/config`. This pollutes the user's git configuration and can cause unexpected fetch behavior for non-checkpoint branches.

## What to implement

Change checkpoint fetch operations to use the remote's URL directly (via `git remote get-url <name>`) instead of the remote name. This ensures git treats the fetch as a one-off operation and doesn't persist refspec configuration.

Additionally, add an integration test that asserts `.git/config` is not modified by checkpoint fetch operations.

## Why this matters

Users reported that after enabling Partio, their `.git/config` accumulated refspec entries like `fetch = +refs/heads/partio/*:refs/remotes/origin/partio/*`. This changes the behavior of `git fetch` for all branches, not just checkpoints, and can confuse users or conflict with other tools that manage git config.

---
id: checkpoint-push-token
target_repos:
  - cli
acceptance_criteria:
  - A new PARTIO_CHECKPOINT_TOKEN environment variable is recognized by the checkpoint push logic
  - When set, the token is used for authentication when pushing/fetching the checkpoint branch to/from a remote
  - The token is injected into the remote URL as an inline credential (https://x-access-token:<token>@github.com/...)
  - Works for both push (pre-push hook) and fetch (partio rewind, partio status)
  - When not set, existing authentication behavior is unchanged
  - The token value is never logged at default log level
  - Unit tests verify URL rewriting and token injection
pr_labels:
  - minion
---

# Add authenticated token for checkpoint push and fetch

Support a `PARTIO_CHECKPOINT_TOKEN` environment variable for authenticating checkpoint branch push and fetch operations, enabling CI/CD pipelines and automated workflows to push checkpoints without interactive git credential setup.

## Context

Partio pushes the checkpoint orphan branch during the pre-push hook. In CI environments, automated pipelines, and remote agent setups, the default git credential chain may not have push access to the checkpoint remote — especially when the checkpoint branch is pushed to a different repository or when the pipeline uses ephemeral credentials.

## Approach

Add support for the `PARTIO_CHECKPOINT_TOKEN` environment variable in the git push and fetch operations used for checkpoint branches. When the token is set:

1. Rewrite the remote URL to inject the token as an inline credential (e.g., `https://x-access-token:<token>@github.com/owner/repo.git`)
2. Use this rewritten URL for `git push` and `git fetch` operations on the checkpoint branch only
3. Never log the token value — mask it in debug output

This follows the same pattern used by GitHub Actions and other CI systems for ephemeral authentication. The token is scoped to checkpoint operations only and does not affect the user's regular git push/pull workflow.

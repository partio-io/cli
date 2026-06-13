---
id: show-descriptive-checkpoint-push-status
target_repos:
  - cli
acceptance_criteria:
  - pre-push hook distinguishes between "pushed new content" and "already up-to-date" when pushing checkpoint branch
  - user-visible stderr message shows "already up-to-date" when no new checkpoint data was transferred
  - user-visible stderr message shows "done" (or similar) only when new content was actually pushed
  - existing behavior preserved when push fails (warn and continue)
pr_labels:
  - minion
---

# Show descriptive checkpoint push status

When Partio pushes the checkpoint branch during `pre-push`, it always prints the same "done" message regardless of whether new data was actually transferred. This makes it impossible for users to tell whether a push was meaningful or a no-op.

## What to implement

Parse the output of `git push` for the checkpoint branch and detect whether content was actually transferred or if the remote was already up-to-date. Show distinct user-facing messages for each case:

- `[partio] Pushing partio/checkpoints/v1… done` — when new checkpoints were pushed
- `[partio] Pushing partio/checkpoints/v1… already up-to-date` — when nothing new to push

## Context hints

- `internal/hooks/prepush.go` — pre-push hook implementation
- `internal/git/push_branch.go` — `PushBranch()` function that executes `git push`

## Why this matters

Users who push frequently see the same message every time and can't tell if their checkpoints are actually being synced. This is especially confusing when debugging checkpoint remote issues — a more descriptive message helps users confirm whether checkpoint data is flowing.

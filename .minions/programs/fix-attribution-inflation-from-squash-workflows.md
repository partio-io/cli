---
id: fix-attribution-inflation-from-squash-workflows
target_repos:
  - cli
acceptance_criteria:
  - attribution calculation excludes intermediate commits that are squashed into the final commit
  - squash-merge workflows produce accurate agent vs human line counts
  - standard (non-squash) commit workflows are unaffected
  - test covers squash scenario where intermediate commits inflate attribution
pr_labels:
  - minion
---

# Fix attribution inflation from squash workflows

When a developer squashes multiple commits into one (e.g., via `git merge --squash` or interactive rebase), intermediate commits that were partially agent-written and partially human-edited can inflate the agent attribution count. The final squashed commit contains only the net changes, but attribution may be calculated against intermediate diffs.

## What to implement

Ensure attribution is calculated against the actual diff of the final commit (parent-to-commit), not accumulated from intermediate commits. When the post-commit hook runs after a squash, it should compare the committed tree against its direct parent — not sum up attribution from the individual pre-squash commits.

Review the attribution calculation in the post-commit flow to verify it uses the final commit's diff rather than any cached or accumulated intermediate state.

## Context hints

- `internal/attribution/` — attribution calculation logic
- `internal/hooks/postcommit.go` — post-commit hook that triggers attribution
- `internal/session/` — session state that may carry intermediate attribution data

## Why this matters

Squash workflows are common in teams that use feature branches. Inflated attribution makes the agent appear responsible for more code than it actually wrote, undermining the accuracy of the audit trail that Partio provides.

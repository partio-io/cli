---
id: sign-checkpoint-branch-bootstrap
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint branch bootstrap commit respects user's commit.gpgsign git config when creating the orphan branch"
  - "Checkpoint branch push succeeds against repos with signed-commit branch protection rulesets"
  - "When GPG/SSH signing is configured, partio uses git commit-tree -S or equivalent to sign the bootstrap commit"
  - "When signing is not configured, behavior is unchanged (unsigned commits still work)"
  - "Unit tests verify signed and unsigned bootstrap paths"
pr_labels:
  - minion
---

# Sign checkpoint branch bootstrap commits to satisfy signed-commit rulesets

Ensure that the initial "bootstrap" commit on the `partio/checkpoints/v1` orphan branch respects the user's `commit.gpgsign` git configuration and any branch protection rulesets requiring signed commits.

Currently, Partio creates orphan branch commits using `git commit-tree` (plumbing), which bypasses porcelain signing behavior. If a repository has GitHub branch protection rulesets requiring signed commits, the unsigned bootstrap commit will be rejected on push.

## Implementation guidance

- When creating the bootstrap commit via `git commit-tree`, check if `commit.gpgsign` is `true` in git config
- If signing is enabled, pass the `-S` flag to `git commit-tree` to produce a signed commit
- Apply the same logic to all subsequent checkpoint commits on the orphan branch
- Consider also checking for `gpg.format` (ssh vs gpg) to ensure compatibility with SSH signing

## Context Hints

- `internal/checkpoint/` — checkpoint storage and git plumbing operations
- `internal/git/` — git operations wrapper

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```

---
id: squash-merge-trailer-preservation
target_repos:
  - cli
acceptance_criteria:
  - "partio doctor warns when the repository's default merge strategy is squash and explains the trailer loss risk"
  - "Documentation or --help text explains the squash merge limitation and recommended workarounds"
  - "If feasible, post-merge hook or pre-push hook detects orphaned checkpoint references and warns the user"
pr_labels:
  - minion
---

# Preserve checkpoint references across squash merges

## Summary

When a branch is squash-merged (GitHub's "Squash and merge" or GitLab's squash option), the original commits — and their `Partio-Checkpoint` trailers — are discarded. The resulting squash commit has no trailer, breaking the link between the merged code and its checkpoint/session data. Partio should detect this scenario and help users preserve or recover checkpoint references.

## Why

Squash merges are a very common workflow (many teams enforce them for clean history). Users who rely on Partio to trace the reasoning behind code changes lose that traceability silently when squash merging. This was reported as a real pain point in entireio/cli#939 (GitLab squash) and is equally relevant for GitHub squash merges. A related issue (entireio/cli#931) shows that fast-forward merges can also lose session references.

## Desired behavior

1. `partio doctor` should detect if the repo's merge strategy commonly drops trailers (e.g., squash merge) and warn about it.
2. Consider adding a post-merge hook that checks whether the merge result lost checkpoint trailers from the merged branch, and if so, copies or re-attaches them to the squash commit.
3. As a fallback, checkpoint data remains on the orphan branch and is still queryable by commit SHA from the original branch.

## Context hints

- `internal/hooks/` — hook implementations
- `cmd/partio/doctor.go` — doctor command for diagnostics
- `internal/checkpoint/` — checkpoint storage and retrieval

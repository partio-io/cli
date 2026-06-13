---
id: compact-commit-trailers
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint trailer value is shortened (e.g., truncated hash or compact format) while remaining unique and resolvable"
  - "partio rewind and other commands that read trailers can resolve both old and new trailer formats"
  - "Multiple checkpoint trailers per commit (if applicable) are consolidated where possible"
  - "Existing commits with old-format trailers continue to work without migration"
pr_labels:
  - minion
---

# Make checkpoint commit trailers more compact

## Summary

The `Partio-Checkpoint` trailer added to commits can be visually noisy in `git log`, GitHub commit views, and PR diffs. When a commit has checkpoint data, the trailer contains a full hash that clutters the commit message. This proposal suggests making the trailer value more compact while keeping it resolvable.

## Why

Community feedback on the analogous feature in entireio/cli (issue #868) highlighted that checkpoint trailers are "too noisy in GitHub commit UI." As Partio adoption grows, users will encounter the same friction — especially in teams where commit messages are reviewed in PRs or changelogs. A more compact trailer format improves developer experience without losing traceability.

## Desired behavior

1. Use a shorter trailer value — for example, a truncated object hash (first 8-12 chars) or a compact reference format.
2. Ensure all Partio commands that read trailers (`rewind`, `status`, etc.) can resolve shortened values.
3. Maintain backward compatibility: old full-hash trailers continue to work.
4. Consider an option to suppress trailers entirely and rely on branch-based checkpoint lookup by commit SHA instead.

## Context hints

- `internal/hooks/post_commit.go` — where trailers are added via `git commit --amend`
- `internal/checkpoint/` — checkpoint storage and ref format
- `cmd/partio/rewind.go` — reads trailers to find checkpoints

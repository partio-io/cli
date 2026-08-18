## Description

Partio's post-commit hook currently performs a full tree walk to determine which files were changed in a commit. This is O(N) in the number of tracked files and becomes a bottleneck on large repos. Replace the tree walk with a `git diff-tree --no-commit-id -r HEAD` invocation, which is a standard plumbing command that returns only the changed files for a given commit. This should noticeably reduce hook latency on repos with many files.

## Source

- **Origin:** entireio/cli#594 (changelog 0.5.0)
- **Detected from:** `entireio-cli`

## Target Repos

- `cli`

## Acceptance Criteria

- [ ] Post-commit hook uses `git diff-tree` instead of a full tree walk to determine changed files
- [ ] Hook produces identical results to the previous implementation for regular commits, merges, and initial commits
- [ ] A benchmark test demonstrates measurably lower latency on a repo with 1000+ files
- [ ] No regression in existing post-commit hook tests

## Context Hints

- `cli/internal/hooks/`
- `cli/internal/git/`

---

Comment `/minion build` or add the `minion-approved` label to begin implementation.

<!-- minion-task
id: post-commit-hook-perf-diff-tree
title: Replace O(N) tree walk with git diff-tree in post-commit hook
source: entireio/cli#594 (changelog 0.5.0)
source_type: changelog
description: Partio's post-commit hook currently performs a full tree walk to determine which files were changed in a commit. This is O(N) in the number of tracked files and becomes a bottleneck on large repos. Replace the tree walk with a `git diff-tree --no-commit-id -r HEAD` invocation, which is a standard plumbing command that returns only the changed files for a given commit. This should noticeably reduce hook latency on repos with many files.
target_repos:
    - cli
context_hints:
    - cli/internal/hooks/
    - cli/internal/git/
acceptance_criteria:
    - Post-commit hook uses `git diff-tree` instead of a full tree walk to determine changed files
    - Hook produces identical results to the previous implementation for regular commits, merges, and initial commits
    - A benchmark test demonstrates measurably lower latency on a repo with 1000+ files
    - No regression in existing post-commit hook tests
pr_labels:
    - minion
    - feature
 minion-task -->

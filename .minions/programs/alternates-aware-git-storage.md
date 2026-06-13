---
id: alternates-aware-git-storage
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint plumbing operations (hash-object, mktree, commit-tree) resolve objects from git alternate object stores
  - partio enable and partio doctor detect alternate object directories and report them in diagnostics
  - Checkpoint creation succeeds in repos with alternates (e.g., CI runners using shared object caches)
  - Existing behavior is unchanged for repos without alternates
pr_labels:
  - minion
---

# Support git alternates in checkpoint plumbing operations

Partio writes checkpoints directly to the git object database using plumbing commands (hash-object, mktree, commit-tree, update-ref). In environments that use git alternates — such as CI runners with shared object caches, `git clone --reference`, or repos linked via `objects/info/alternates` — object resolution may fail because Partio's plumbing operations don't account for alternate object directories.

## What to implement

Update the checkpoint storage layer (`internal/checkpoint/`) to open git repositories with alternates-aware object resolution:

1. When resolving git objects for checkpoint operations, check `objects/info/alternates` for additional object store paths.
2. Ensure `commit-tree` can reference parent commits that live in alternate object stores.
3. Add alternates detection to `partio doctor` diagnostics so users can see if alternates are configured.
4. Add test coverage for checkpoint creation in a repo with an alternates file pointing to a shared object store.

## Why this matters

CI/CD environments commonly use git alternates to reduce clone times and disk usage. If Partio can't create checkpoints in these environments, users lose session capture for their CI-driven agent workflows. This is increasingly important as more teams run AI coding agents in CI pipelines.

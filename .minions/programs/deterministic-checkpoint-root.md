---
id: deterministic-checkpoint-root
target_repos:
  - cli
acceptance_criteria:
  - "The orphan checkpoint branch has a deterministic root commit with a fixed tree and message"
  - "Creating the root commit is idempotent — if one already exists, it is reused"
  - "Existing checkpoint branches with a non-deterministic root continue to work without migration"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add deterministic root anchor commit for checkpoint branch

When Partio creates the orphan `partio/checkpoints/v1` branch, the initial commit should use a deterministic tree hash and commit message so that independent clones or worktrees that initialize the branch produce the same root object. This makes it possible to fast-forward merge checkpoint branches from different machines without conflicts at the root.

## Context

Currently the checkpoint branch is initialized with whatever state happens to be first. If two machines independently initialize the branch, they get divergent roots and cannot be reconciled without a force-push. A fixed "Initialize checkpoint branch" root commit with an empty tree eliminates this class of conflict.

## Implementation hints

- In the checkpoint storage layer (`internal/checkpoint/`), when creating the orphan branch for the first time, use `git mktree` with an empty input to produce the well-known empty tree hash, then `git commit-tree` with a fixed message and zero timestamp.
- Before creating a new root, check if the branch ref already exists and skip initialization if so.
- Add tests verifying that two independent initializations produce the same commit hash.

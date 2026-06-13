---
id: sha256-git-object-format-support
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint write operations (hash-object, mktree, commit-tree, update-ref) succeed in SHA-256 repos"
  - "Existing SHA-1 repo behavior is unchanged"
  - "partio doctor detects and reports the repo's object format"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Support SHA-256 git object format in checkpoint operations

Partio uses git plumbing commands (hash-object, mktree, commit-tree, update-ref) to write checkpoints to an orphan branch. These operations may fail or produce invalid refs in repositories configured with SHA-256 object format (`git init --object-format=sha256`).

## What to implement

1. Detect the repository's object format via `git rev-parse --show-object-format` and pass it through to checkpoint operations where needed.
2. Ensure all git plumbing calls that create or reference objects handle both SHA-1 (40 hex chars) and SHA-256 (64 hex chars) hashes correctly.
3. Add object format detection to `partio doctor` output so users can see which format their repo uses.
4. Add test coverage for SHA-256 repos (use `git init --object-format=sha256` in test setup where applicable).

## Context hints

- `internal/checkpoint/` — checkpoint storage using git plumbing
- `internal/git/` — git operations wrapper
- `cmd/partio/doctor.go` — doctor command

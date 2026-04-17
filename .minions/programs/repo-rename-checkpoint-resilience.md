---
id: repo-rename-checkpoint-resilience
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint references use repo-local identifiers that survive repository renames"
  - "Existing checkpoint data remains accessible after a GitHub repository rename"
  - "The `partio doctor` command detects and warns about stale remote URLs after a rename"
  - "Documentation or help text explains how to update checkpoint remotes after a rename"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Ensure checkpoint linkage survives repository renames

Make checkpoint data resilient to GitHub repository renames so that users don't lose session linkage when renaming their repo.

## What to implement

1. **Repo-local checkpoint references** — ensure checkpoint branch references and metadata use repo-local identifiers (commit hashes, branch names) rather than absolute URLs that break on rename
2. **Doctor check** — add a diagnostic to `partio doctor` that detects when the configured checkpoint remote URL no longer matches the current origin (which happens after a rename if git's redirect stops working)
3. **Remediation guidance** — when the doctor check detects a stale URL, provide clear instructions for updating the remote

## Why this matters

Repository renames are common (rebranding, org transfers, correcting typos). When a repo is renamed on GitHub, git's HTTP redirect handles pushes/fetches temporarily, but the old URL eventually stops resolving. If checkpoint data references the old repo name, sessions become unlinkable. This was flagged as a user concern in entireio/cli#909.

## Context hints

- `internal/checkpoint/` — checkpoint storage and references
- `internal/git/` — git remote operations
- `cmd/partio/doctor.go` — doctor command diagnostics

---
id: fetch-remote-checkpoint-on-enable
target_repos:
  - cli
acceptance_criteria:
  - partio enable fetches existing partio/checkpoints/v1 from origin if the branch exists remotely but not locally
  - If the branch exists both locally and remotely, partio enable warns but does not overwrite the local branch
  - If the branch exists only locally, partio enable proceeds as before (no remote fetch)
  - If the remote branch does not exist, partio enable creates a new orphan branch as before
  - partio doctor reports when local and remote checkpoint branches have diverged
pr_labels:
  - minion
---

# Fetch existing remote checkpoint branch on enable

## Summary

When `partio enable` runs in a repo that already has a `partio/checkpoints/v1` branch on the remote (e.g. set up on another device), fetch that branch instead of creating a new empty orphan branch.

## Context

In multi-device workflows, a user enables Partio on device A, creates checkpoints, and pushes them. When they clone the repo on device B and run `partio enable`, Partio currently creates a fresh empty orphan branch, discarding the existing checkpoint history. A subsequent fetch is rejected as non-fast-forward because the histories are unrelated.

Inspired by entireio/cli#1374 which reported the same orphan-branch conflict in multi-device setups.

## Approach

- During `partio enable`, before creating the orphan branch, check if `partio/checkpoints/v1` exists on the remote using `git ls-remote`
- If it exists remotely but not locally, fetch it: `git fetch origin partio/checkpoints/v1:partio/checkpoints/v1`
- If it exists both locally and remotely, warn the user about potential divergence and suggest `partio doctor` for diagnosis
- Add a divergence check to `partio doctor` that compares local and remote checkpoint branch tips
- This aligns with the existing git plumbing approach and doesn't require new dependencies

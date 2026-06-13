---
id: custom-ref-checkpoint-storage
target_repos:
  - cli
acceptance_criteria:
  - Add a `checkpoints_ref` config option (e.g., `custom`) that stores checkpoints at `refs/partio/checkpoints/v1` instead of `refs/heads/partio/checkpoints/v1`
  - When `checkpoints_ref` is set to `custom`, all checkpoint reads/writes use the custom ref namespace
  - Default behavior (`checkpoints_ref` unset or `branch`) remains unchanged for backwards compatibility
  - `partio enable` installs a fetch refspec (`+refs/partio/checkpoints/v1:refs/partio/remotes/origin/checkpoints/v1`) when custom ref mode is active
  - Checkpoints stored at custom refs are invisible to `git branch -a` and GitHub branch UI
  - Pre-push hook pushes the custom ref correctly
  - `partio doctor` validates the configured ref location
  - Migration path: when switching from branch to custom ref, preserve existing checkpoint history by pointing the custom ref at the branch tip
  - All existing commands (status, clean, rewind, reset) work with either ref location
pr_labels:
  - minion
---

# Store checkpoint metadata at custom git refs

## Summary

Add an option to store Partio checkpoint metadata at a custom git ref (`refs/partio/checkpoints/v1`) instead of a branch (`refs/heads/partio/checkpoints/v1`). This hides checkpoint data from `git branch -a`, GitHub's branch UI, and default `git clone` fetches, reducing noise for users who find the checkpoint branch confusing or cluttering.

## Motivation

Currently, Partio stores all checkpoint data on an orphan branch named `partio/checkpoints/v1`. While functional, this has drawbacks:

1. **Branch list noise**: The checkpoint branch appears in `git branch -a` and GitHub's branch dropdown, confusing collaborators who don't use Partio.
2. **Default clone bloat**: `git clone` fetches the checkpoint branch by default, adding download time for repos with large checkpoint histories.
3. **UI confusion**: GitHub shows the checkpoint branch in the branch count and branch picker, which can be misleading.

Using a custom ref namespace (`refs/partio/checkpoints/v1`) solves all three issues while keeping the exact same on-disk format and sharded path structure.

## Implementation Notes

- The checkpoint ref path should be configurable via `checkpoints_ref` in settings (values: `branch` (default) or `custom`).
- All code that currently references the hardcoded `partio/checkpoints/v1` branch (in `internal/checkpoint/store.go`, `internal/git/git.go`, `cmd/partio/*.go`) should resolve the ref through a function that respects the setting.
- The `update-ref` calls in checkpoint storage already work with arbitrary refs, so the plumbing layer needs minimal changes.
- Push hook should use the resolved ref when pushing to origin.
- A one-time migration preserves history: `git update-ref refs/partio/checkpoints/v1 refs/heads/partio/checkpoints/v1`.

## Source

Inspired by entireio/cli#1242 (Checkpoints v1.1: store v1 metadata at a custom ref).

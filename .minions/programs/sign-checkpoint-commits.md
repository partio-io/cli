---
id: sign-checkpoint-commits
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint commits on the orphan branch are signed using the user's configured git signing key (GPG or SSH) when available
  - Signing is best-effort — if no signing key is configured or the signer fails, the commit is created unsigned and a warning is logged
  - A configuration option `sign_checkpoint_commits` (default true) allows users to opt out of checkpoint signing
  - Unit tests verify that the signing configuration is read and passed to commit creation, and that failures are handled gracefully
pr_labels:
  - minion
---

# Sign checkpoint commits with user's git signing key

## Problem

Partio creates checkpoint commits on an orphan branch using git plumbing commands (`commit-tree`). These commits are currently always unsigned, even when the user has configured git commit signing (`commit.gpgsign=true`). This creates an inconsistency: the user's regular commits are signed, but the checkpoint commits that Partio creates alongside them are not.

In environments with signing requirements (corporate policies, supply chain security), unsigned checkpoint commits may trigger warnings or be rejected by server-side hooks on push.

## Proposed solution

Extend the checkpoint commit creation path to optionally sign commits:

1. **Detect signing config**: Read `commit.gpgsign`, `user.signingkey`, and `gpg.format` from git config to determine if signing is enabled and which key/format to use.
2. **Sign via git plumbing**: Use `git commit-tree -S` (or `-S<key>`) when creating checkpoint commits. This works with both GPG and SSH signing.
3. **Best-effort**: If signing fails (missing key, locked agent, etc.), fall back to creating an unsigned commit and log a warning. Checkpoint creation must never fail due to signing issues.
4. **Configuration**: Add `sign_checkpoint_commits` boolean to settings (default: `true`). When `false`, Partio skips signing even if the user has `commit.gpgsign=true` globally — useful for CI environments where signing infrastructure isn't available.

## Why this matters

As more organizations adopt commit signing for supply chain security, having unsigned commits in the checkpoint branch creates friction. Best-effort signing aligns Partio's checkpoint commits with the user's existing signing configuration without adding setup burden or risking checkpoint creation failures.

## Context hints

- `internal/checkpoint/` — checkpoint storage and commit creation via git plumbing
- `internal/git/` — git operations layer
- `internal/config/` — configuration layer for the new setting

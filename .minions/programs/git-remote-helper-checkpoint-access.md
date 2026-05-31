---
id: git-remote-helper-checkpoint-access
target_repos:
  - cli
acceptance_criteria:
  - A `git-remote-partio` binary is built alongside the main CLI
  - Users can add a partio remote (`git remote add checkpoints partio://repo`) to fetch checkpoint refs
  - The remote helper resolves checkpoint branch refs and serves objects from the local or configured checkpoint remote
  - Integration test demonstrates clone/fetch of checkpoint data via the helper protocol
pr_labels:
  - minion
  - enhancement
---

# Add git remote helper for checkpoint access

Implement a custom git remote helper (`git-remote-partio`) that allows users to access checkpoint data using standard git remote protocols.

## Motivation

Currently, checkpoint data lives on an orphan branch that users interact with indirectly through Partio commands. A git remote helper would let users browse, fetch, and share checkpoint branches using familiar git tooling (`git fetch`, `git log`, etc.) without needing to know the internal ref layout.

Inspired by entireio/cli PR #1306 which ships a `git-remote-entire` binary for transparent checkpoint access via `git clone entire://...`.

## Implementation Notes

- Add a `cmd/git-remote-partio/` binary that implements the git remote helper protocol (capabilities: fetch, list)
- The helper should resolve `partio://` URLs to the configured checkpoint remote or local refs
- Build via goreleaser as a secondary binary alongside the main `partio` CLI
- Keep the binary lean — only checkpoint ref resolution, no full CLI dependency

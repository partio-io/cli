---
id: configurable-trailer-visibility
target_repos:
  - cli
acceptance_criteria:
  - "New `trailer_mode` config option in settings.json with values: `full` (default, current behavior), `short` (abbreviated hash only), `hidden` (no trailer added)"
  - "`full` mode writes the current `Partio-Checkpoint: <full-hash>` trailer"
  - "`short` mode writes `Partio-Checkpoint: <8-char-hash>` for a less noisy commit message"
  - "`hidden` mode skips trailer insertion entirely — checkpoint is still created and linked via internal metadata only"
  - "Setting is configurable via `partio configure --trailer-mode <mode>`"
  - "`partio doctor` warns when `hidden` mode is active, explaining that checkpoint linkage depends on internal state only"
  - "Config is read from the layered config system (defaults → global → repo → local → env)"
pr_labels:
  - minion
---

# Add configurable checkpoint trailer visibility

## Description

Add a `trailer_mode` configuration option that controls how the `Partio-Checkpoint` trailer appears in commit messages. Currently, the trailer is always written in full, which some users find noisy in GitHub's commit UI. This feature offers three modes:

- **`full`** (default): Current behavior — full checkpoint hash in the trailer
- **`short`**: Abbreviated 8-character hash — less visual noise while preserving linkage
- **`hidden`**: No trailer at all — checkpoint is still created on the orphan branch, but linkage relies on commit timestamp and tree hash correlation rather than an explicit trailer

The setting should be part of the layered config system and configurable via the CLI.

## Why

The checkpoint trailer is the durable link between a user commit and its checkpoint metadata. However, for teams that find it visually noisy (especially in GitHub's commit list UI), there's currently no way to reduce its footprint. Offering a `short` mode preserves linkage with less noise, while `hidden` mode serves teams that prioritize clean commit messages and are willing to rely on heuristic linkage.

## Source

Inspired by community feedback on entireio/cli issue #868 about the `Entire-Checkpoint` trailer being too noisy in GitHub's commit UI.

## Context hints

- `internal/hooks/` — post-commit hook that adds the trailer via `git commit --amend`
- `internal/config/` — layered configuration system
- `cmd/partio/` — CLI commands including `configure`

---
id: auto-update-cli
target_repos:
  - cli
acceptance_criteria:
  - Global settings file supports an `auto_update` field with values `off` (default), `prompt`, or `auto`
  - When a newer version is available, `partio` displays a notification with update instructions
  - In `prompt` mode, partio asks Y/N before updating; in `auto` mode it updates silently
  - `partio update [--check-only]` command runs the installer on demand
  - Install provenance (method, timestamp) is recorded in the global config directory
  - Dev builds skip update checks entirely
pr_labels:
  - minion
---

# Auto-update CLI behind opt-in global setting

## Description

Add an opt-in auto-update mechanism to the Partio CLI. When enabled via global settings (`~/.config/partio/settings.json`), the CLI checks for newer releases and either notifies, prompts, or silently updates depending on the configured mode.

A new `partio update [--check-only]` command provides on-demand update capability regardless of the auto-update setting.

## Why

As Partio evolves — especially hook formats and checkpoint schemas — users running stale CLI versions encounter cryptic failures. An auto-update mechanism reduces support burden and ensures users benefit from bug fixes and new features without manual intervention.

## Context hints

- `internal/config/` — global settings loading
- `cmd/partio/` — CLI command registration
- `internal/config/defaults.go` — default config values

## Source

Inspired by entireio/cli PR #981 (auto-update feature behind opt-in global setting).

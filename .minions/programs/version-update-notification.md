---
id: version-update-notification
target_repos:
  - cli
acceptance_criteria:
  - "When a newer version is available, partio prints a one-line notice to stderr after command output"
  - "Version check is non-blocking and does not slow down CLI commands"
  - "Check frequency is rate-limited (e.g., once per 24 hours) using a local cache file"
  - "Notification can be suppressed via environment variable (e.g., PARTIO_NO_UPDATE_CHECK=1)"
  - "Version check works for GitHub releases (no external service dependency)"
pr_labels:
  - minion
---

# Notify users when a newer partio version is available

## Summary

Add a lightweight version check that notifies users when a newer version of the Partio CLI is available. The check should be non-intrusive — a single line printed to stderr after normal command output, rate-limited to avoid repeated noise.

## Why

Partio is actively developed and users on older versions may miss important bug fixes or features. The entireio/cli project added inline auto-update prompts in v0.5.6 and aligned them across installers in v0.6.0, indicating this is a proven UX pattern for CLI tools in this space. Unlike auto-update, a notification-only approach is safe and non-disruptive.

## Desired behavior

1. After any `partio` command completes, check if a newer GitHub release exists (cached, non-blocking).
2. If a newer version is found, print a single line to stderr: `A newer version of partio is available (vX.Y.Z). Run: go install github.com/partio-io/cli/cmd/partio@latest`
3. Cache the check result locally (e.g., `~/.config/partio/version-check.json`) with a TTL of 24 hours.
4. Skip the check if `PARTIO_NO_UPDATE_CHECK=1` is set, or if running in CI (detect `CI=true`).
5. The check must not add latency to commands — run in a background goroutine or check only cached data.

## Context hints

- `cmd/partio/root.go` — root command setup where a post-run hook could trigger the check
- `internal/config/` — config/cache file location conventions

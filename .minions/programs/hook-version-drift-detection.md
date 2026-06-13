---
id: hook-version-drift-detection
target_repos:
  - cli
acceptance_criteria:
  - Installed hook scripts contain a comment of the form `# partio-version: <semver>` written at install time
  - "`partio status` reads the version from all three hook scripts and warns if any differ from the running binary version"
  - "`partio doctor` reports a WARN item when installed hook version does not match the binary version"
  - "Hooks that predate this feature (missing the version comment) are reported as 'unknown version' rather than causing a panic or silent pass"
  - "Existing unit tests in hooks_test.go continue to pass; new tests verify version embedding and drift detection logic"
pr_labels:
  - minion
---

# Stamp hook scripts with CLI version and warn on drift

Embed the partio CLI version into installed hook scripts (e.g., as a comment `# partio-version: v1.2.3` alongside the existing `# Installed by partio` marker). In `partio status`, parse the installed hook scripts to extract this version and compare it against the running binary version. If they differ, print a warning such as `Hooks: installed (outdated, run 'partio enable' to upgrade)`. The `partio doctor` command should also surface this as a diagnostic.

## Why

Users often upgrade the partio binary but forget to reinstall hooks, leading to subtle mismatches where older hook behavior persists silently. Surfacing stale hooks in `status` and `doctor` gives users clear, actionable feedback.

## User relevance

Prevents silent failures where a user has upgraded partio but their git hooks still invoke old behavior. The drift warning in `partio status` makes the upgrade path obvious without requiring users to know the internal hook format.

## Context hints

- `internal/git/hooks/hooks.go`
- `internal/git/hooks/install.go`
- `cmd/partio/status.go`
- `cmd/partio/doctor.go`

## Source

Inspired by entireio/cli#982 (stamp agent hook configs with CLI version for drift detection).

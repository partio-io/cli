---
id: enrich-checkpoint-cli-version
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint metadata cli_version field includes the git tag if built from a tagged commit
  - Checkpoint metadata cli_version field includes a dirty flag if built with uncommitted changes
  - Version string format is consistent (e.g. "v0.3.0", "v0.3.0-dirty", "v0.3.0-4-gabcdef", "v0.3.0-4-gabcdef-dirty")
  - Existing version embedding via ldflags continues to work
  - partio doctor displays the enriched version string
pr_labels:
  - minion
---

# Enrich checkpoint cli_version with tag and dirty state

## Summary

Include the full git-describe output (tag + distance + commit + dirty flag) in the `cli_version` field stored in checkpoint metadata, rather than just the bare version number.

## Context

When debugging checkpoint issues or auditing which CLI version produced a checkpoint, knowing the exact build state matters. A bare "0.3.0" doesn't distinguish between a clean tagged release and a local dev build with uncommitted changes. This is especially important when users build from source or run pre-release versions.

Inspired by entireio/cli#1275 which enriched their checkpoint cli_version with tag and dirty state.

## Approach

- Update the version embedding in the Makefile to use `git describe --tags --dirty --always` instead of just the tag
- Ensure the `version` variable embedded via `-ldflags` carries the full describe string
- The checkpoint metadata writer already stores `cli_version` from this variable, so no changes needed there
- Update `partio doctor` to display the enriched version for diagnostics
- Fall back to Go build info (`debug.ReadBuildInfo`) if ldflags version is empty (e.g. `go install` without `-ldflags`)

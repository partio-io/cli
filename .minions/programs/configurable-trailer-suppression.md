---
id: configurable-trailer-suppression
target_repos:
  - cli
acceptance_criteria:
  - A new `trailers` config section supports `enabled: false` to skip amending trailers entirely
  - A `key` override in config causes the checkpoint trailer to use the custom key instead of `Partio-Checkpoint`
  - When `trailers.enabled` is false, `runPostCommit` skips calling `git.AmendTrailers` but still writes the checkpoint
  - The existing default behavior (trailers on) is preserved when no config override is present
  - Table-driven tests cover enabled/disabled/custom-key scenarios
pr_labels:
  - minion
---

# Configurable commit trailer suppression or renaming

## Description

Add config options to allow users to suppress the Partio-Checkpoint and Partio-Attribution trailers entirely, or rename the trailer keys. This should be controllable via `.partio/settings.json` (e.g. `trailer_key`, `include_trailers: false`). The `AmendTrailers` function and the trailer construction in `runPostCommit` should read these config values.

## Why

Commit trailers are visible in GitHub's UI and in `git log`, which can feel noisy or expose internal tooling details to external contributors. Some teams want the checkpoint stored silently without polluting commit messages.

## User Relevance

Users working in public or team repos who want Partio to capture sessions invisibly, without advertising the tooling in every commit message.

## Context Hints

- `internal/git/amend_trailers.go`
- `internal/hooks/postcommit.go`
- `internal/config/config.go`
- `internal/config/defaults.go`

## Source

entireio/cli#1084, entireio/cli#868

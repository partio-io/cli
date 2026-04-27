---
id: settings-schema-validation
target_repos:
  - cli
acceptance_criteria:
  - Config loading returns a typed validation error when unknown fields are present in settings JSON
  - Config loading rejects invalid enum values for strategy and log_level with a clear message
  - partio doctor reports configuration validation warnings
  - Existing valid configs continue to load without errors (backwards compatible)
pr_labels:
  - minion
---

# Add structured validation to settings configuration

## Summary

Add a validation layer to the config loading pipeline that catches misconfigurations early with clear, actionable error messages. Currently, `internal/config/load.go` deserializes JSON settings files without validating field names, value ranges, or structural correctness — typos and invalid values are silently accepted.

## What to implement

1. Add a `Validate() error` method to `Config` in `internal/config/config.go` that checks:
   - `strategy` is one of the known strategy values (or empty for default)
   - `log_level` is one of: debug, info, warn, error (or empty for default)
   - `agent` is a recognized agent name (or empty for default)
   - No unknown top-level keys are present (use `json.Decoder.DisallowUnknownFields` or manual key checking)

2. Call `Validate()` at the end of `Load()` in `internal/config/load.go`, after all layers are merged.

3. Add a validation check to `partio doctor` that loads config and reports any validation warnings.

## Context

- `internal/config/config.go` — Config struct definition
- `internal/config/load.go` — layered config loading
- `internal/config/defaults.go` — default values
- `cmd/partio/doctor.go` — doctor command

## Why

Silent misconfiguration is a common source of confusion — users may set `log_level: "verbose"` (invalid) or `stratgey` (typo) and wonder why their settings have no effect. Validation catches these at load time rather than at runtime or not at all. Inspired by entireio/cli's settings schema v2 (PR #1051) which adds typed validation on every config load.

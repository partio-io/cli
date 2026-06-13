---
id: doctor-preflight-ci-integration
target_repos:
  - cli
acceptance_criteria:
  - partio doctor exits with non-zero status when actionable issues are found (e.g., hooks out of date, config invalid)
  - partio doctor supports a --ci or --json flag that outputs machine-readable results suitable for CI pipelines
  - Doctor checks include hook version drift detection (hooks installed by older CLI version)
  - Tests verify exit codes for healthy and unhealthy configurations
pr_labels:
  - minion
---

# Add CI-friendly output mode to partio doctor for automated preflight checks

`partio doctor` currently runs interactive diagnostics, but there is no clean way to use it as an automated gate in CI pipelines or test harnesses. Teams using Partio in CI environments need a way to verify that hook configurations are valid and up-to-date before running agent sessions.

## What to implement

1. Add a `--ci` flag (or `--json`) to `partio doctor` that:
   - Outputs structured results (JSON or one-line-per-check format)
   - Exits with a non-zero status code when actionable issues are found
   - Suppresses interactive formatting (colors, spinners)
2. Ensure doctor checks cover:
   - Hook installation status and version drift (hooks installed by an older CLI version)
   - Settings file validity
   - Agent configuration consistency
3. Document the exit codes so CI scripts can distinguish between "all good", "warnings only", and "blocking issues".

## Why this matters

As Partio adoption grows in teams, catching configuration drift early in CI prevents silent checkpoint failures that only surface when developers notice missing trailers on their commits. Running `partio doctor` as a CI preflight closes this gap.

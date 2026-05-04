---
id: session-driven-config-recommendations
target_repos:
  - cli
acceptance_criteria:
  - A `partio improve` command analyzes completed session checkpoints and outputs actionable configuration recommendations
  - Recommendations include specific settings changes with rationale based on observed session patterns
  - The command outputs structured JSON with --json flag for machine consumption
  - Recommendations are based on concrete metrics (e.g., session duration, checkpoint frequency, hook timing, error rates) not heuristics alone
pr_labels:
  - minion
---

# Add session-driven configuration recommendations

## Summary

Add a `partio improve` command that analyzes historical session checkpoints and generates actionable configuration recommendations. Unlike `partio insights` (pattern analysis and reporting), this command produces specific, implementable suggestions for improving Partio's configuration based on observed session behavior.

## Why

Users configure Partio once during setup and rarely revisit their configuration. Over time, usage patterns emerge that suggest better settings — for example, if hook timeouts are consistently near the limit, the timeout should be increased; if sessions are very long, transcript condensation thresholds should be adjusted; if checkpoint push failures correlate with large transcript sizes, redaction or compaction settings may help. Currently users must manually correlate these patterns. An automated recommendation engine turns checkpoint data into actionable improvements.

## What to implement

1. Add a `partio improve` command that reads checkpoint metadata and session history from the checkpoint branch.
2. Implement analysis rules that detect common configuration improvement opportunities:
   - Hook timing patterns suggesting timeout adjustments
   - Session size patterns suggesting condensation or retention changes
   - Error frequency patterns suggesting resilience configuration
   - Agent detection patterns suggesting detector configuration tuning
3. Output recommendations as a prioritized list with:
   - The specific setting to change and its recommended value
   - The evidence from session data supporting the recommendation
   - The expected impact of the change
4. Support `--json` for machine-readable output and `--apply` to interactively apply selected recommendations.

## Context hints

- `internal/checkpoint/` — checkpoint storage and reading
- `internal/session/` — session data
- `internal/config/` — configuration system
- `cmd/partio/` — command definitions

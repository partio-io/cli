---
id: add-token-usage-comparison-between-checkpoints
target_repos:
  - cli
acceptance_criteria:
  - "partio status or a new subcommand can show token usage delta between two checkpoints"
  - "Comparison shows input tokens, output tokens, and cache token deltas"
  - "JSON output mode is supported for machine consumption"
  - "Graceful handling when token data is missing from one or both checkpoints"
  - "make test passes with new test cases"
  - "make lint passes"
pr_labels:
  - minion
---

# Add token usage comparison between checkpoints

Once token usage metrics are tracked in checkpoint metadata (#333), users need a way to compare token consumption between checkpoints to understand how token usage evolves across a session. This helps identify which steps in an agent session were most expensive and supports cost optimization.

## What to implement

Add a comparison capability that shows the delta in token usage between two checkpoints:
- Input tokens delta
- Output tokens delta
- Cache read/write token deltas (if available)
- Total token delta

This could be a flag on an existing command (e.g. `partio checkpoint compare <id1> <id2>`) or integrated into status output. The implementation should support `--json` output for machine consumption.

### Dependency

This feature depends on #333 (Track token usage metrics in checkpoint metadata) being implemented first, since it needs token data in checkpoint metadata to compare.

### Where to look
- `internal/checkpoint/` — checkpoint domain type and storage
- `cmd/partio/` — CLI command definitions

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```

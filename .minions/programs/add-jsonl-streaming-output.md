---
id: add-jsonl-streaming-output
target_repos:
  - cli
acceptance_criteria:
  - partio status --output jsonl emits one JSON object per line to stdout
  - partio rewind --output jsonl emits structured progress events as JSONL
  - JSONL output mode suppresses all human-oriented formatting (spinners, colors, tables)
  - Each JSONL line includes a "type" field identifying the event kind (e.g. "status", "checkpoint", "error")
  - Invalid --output values produce a clear error message
  - Existing default output behavior is unchanged when --output is not specified
pr_labels:
  - minion
---

# Add `--output jsonl` streaming mode for machine-readable CLI output

Add a global `--output` flag (starting with `jsonl` format) so that CLI commands can emit structured, machine-readable events as newline-delimited JSON. This enables piping Partio output into other tools, dashboards, and CI pipelines.

## Motivation

Entire CLI v0.6.2 added JSONL output modes to `entire review`, allowing live streaming of agent events for programmatic consumption. Partio currently only outputs human-readable text, which is difficult to parse in automation. A structured output mode would make Partio composable with other tools — especially useful for CI integration, custom dashboards, and scripting workflows.

## Implementation notes

- Add a `--output` persistent flag to the root command (default: empty string for human output)
- Support `jsonl` as the first format; design for future formats (e.g., `json` for single-object output)
- Each command that supports it should emit typed JSONL events with at minimum `{"type": "...", "data": {...}}`
- Start with `partio status` and `partio rewind` as initial commands
- When `--output jsonl` is active, suppress all non-structured output (progress spinners, color codes, tables)

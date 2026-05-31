---
id: hook-execution-tracing-spans
target_repos:
  - cli
acceptance_criteria:
  - Hook execution emits structured timing spans (start/end/duration) for each phase
  - Spans cover at minimum: agent detection, session discovery, JSONL parsing, checkpoint creation, ref update
  - Timing data is logged at debug level and available via `partio doctor --timing`
  - Total hook latency is reported in status output when it exceeds a configurable threshold
  - No measurable performance overhead when logging is at info level or above
pr_labels:
  - minion
  - enhancement
---

# Add nested performance spans for hook execution observability

Implement structured tracing spans within hook execution so users can diagnose slow hooks and understand where time is spent.

## Motivation

Hook latency directly impacts developer experience — slow hooks block commits. Users report hook timeouts (entireio/cli issues #1137, #1072) but lack visibility into which phase is slow. Adding lightweight tracing spans lets users and maintainers pinpoint bottlenecks (is it agent detection? JSONL parsing? git plumbing?) without guessing.

Inspired by entireio/cli changelog 0.6.0 (nested performance spans in traces) and issues #1137/#1072 (hook timeout reports).

## Implementation Notes

- Add a lightweight `internal/trace` package with `StartSpan(name)` / `span.End()` API
- Spans nest naturally via context or explicit parent — no external tracing dependency needed
- In debug mode, spans emit structured log lines: `span_name duration_ms parent_span`
- Add a `partio doctor --timing` flag that runs a synthetic hook cycle and reports span breakdown
- Store last hook timing in `.partio/state/last-timing.json` for `partio status` to display warnings

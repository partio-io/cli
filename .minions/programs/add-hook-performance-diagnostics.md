---
id: add-hook-performance-diagnostics
target_repos:
  - cli
acceptance_criteria:
  - Each hook invocation records its wall-clock execution time
  - partio doctor displays hook timing from the most recent commit
  - Hook timing is logged at debug level during normal operation
  - When any hook exceeds a configurable threshold (default 5s), a warning is emitted
  - partio status --verbose shows aggregate hook timing statistics
  - Timing data is stored in .partio/state/ and cleaned up with other state files
pr_labels:
  - minion
---

# Add hook execution timing and performance diagnostics

Record and surface hook execution timing to help users diagnose latency issues caused by Partio's git hooks.

## Motivation

Multiple reports in the Entire CLI ecosystem (entireio/cli#1137, #1072, #957) describe hook timeouts and latency concerns — hooks adding 30+ seconds to agent responses. While Partio has a fast-fail startup timeout proposal (#343), there is no way for users to understand *how long* each hook phase takes or identify which part of the checkpoint pipeline is slow. Visibility into hook performance is essential for diagnosing and fixing latency complaints.

## Implementation notes

- In the hook implementations (`internal/hooks/`), wrap each major phase (agent detection, state save/load, attribution calculation, JSONL parsing, checkpoint creation, commit amend) with timing measurements using `time.Now()` / `time.Since()`
- Write a timing summary to `.partio/state/hook-timing.json` after each hook completes, structured as `{"hook": "post-commit", "phases": [{"name": "detect_agent", "duration_ms": 42}, ...], "total_ms": 350}`
- Surface timing in `partio doctor` by reading the most recent timing file
- Add a `hook_warn_threshold_ms` config option (default: 5000) that triggers a warning when any hook exceeds the threshold
- Log all timing at debug level so `PARTIO_LOG_LEVEL=debug` gives full visibility
- Clean up timing files alongside other state files in the existing cleanup flow

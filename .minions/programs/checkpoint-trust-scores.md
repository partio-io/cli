---
id: checkpoint-trust-scores
target_repos:
  - cli
acceptance_criteria:
  - "Each checkpoint includes a trust score in its metadata"
  - "Trust score is computed from observable signals: session length, error count, test results presence"
  - "Score is stored in metadata.json alongside existing attribution data"
  - "`partio status` displays the trust score when available"
  - "Trust score computation does not block git operations"
pr_labels:
  - minion
---

# Add trust scores to checkpoint metadata

## Summary

Compute and store a trust/confidence score alongside each checkpoint's metadata. The score provides a quick signal about how reliable a particular agent-assisted commit is, based on observable session characteristics like session completeness, error frequency, and whether tests were run.

## Motivation

Not all agent sessions are equal — some are clean, focused implementations with tests, while others involve extensive debugging or incomplete work. A trust score gives reviewers an at-a-glance signal about the quality of agent-assisted commits. This is especially valuable in team settings where multiple developers use AI agents and reviewers need to prioritize which commits deserve closer inspection.

## Behavior

1. After checkpoint creation, compute a trust score (0.0-1.0) from:
   - Session completeness (did the agent finish vs. get interrupted?)
   - Error/retry frequency in the transcript
   - Presence of test execution in the session
   - Session duration relative to change complexity
2. Store the score in `metadata.json` under a `trust_score` field
3. `partio status` shows the score (e.g., `trust: 0.85`)
4. Score computation is best-effort and non-blocking

## Context

- Inspired by `entireio/cli` issue #790 (trust scores alongside checkpoint context)
- `internal/checkpoint/` for metadata storage
- `internal/attribution/` for existing quality signal computation

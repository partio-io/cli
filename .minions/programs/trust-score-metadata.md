---
id: trust-score-metadata
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint metadata includes a structured trust_score object with source, verification_status, and confidence fields
  - Trust score is populated from agent transcript analysis (e.g., whether the agent cited sources, ran tests, or expressed uncertainty)
  - partio status and checkpoint inspect commands display trust score when present
  - Trust score defaults to null/omitted when no signal is available (no false confidence)
pr_labels:
  - minion
---

# Add trust/confidence metadata to checkpoints

## Summary

Checkpoints capture what the agent thought and decided, but there's no structured way to assess **how trustworthy** the resulting code is. As agents generate more code at scale, reviewers need a quick signal beyond reading the full transcript.

## What to implement

Add an optional `trust_score` field to checkpoint metadata containing:
- `sources_cited`: whether the agent referenced documentation, tests, or existing code
- `tests_executed`: whether the agent ran tests and they passed
- `uncertainty_signals`: count of hedging language or explicit uncertainty in the transcript
- `confidence`: a simple low/medium/high enum derived from the above signals

The score should be computed during checkpoint creation by analyzing the session transcript. It should be displayed in `partio status` and any checkpoint inspection commands.

## Why this matters

Reviewers currently read the full transcript to assess trust. This doesn't scale when agents produce dozens of commits per day. A structured trust signal lets teams triage reviews — spending more time on low-confidence changes and fast-tracking high-confidence ones.

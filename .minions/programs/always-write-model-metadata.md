---
id: always-write-model-metadata
target_repos:
  - cli
acceptance_criteria:
  - "Checkpoint metadata always includes a 'model' field, even when empty string"
  - "The model JSON tag does not use omitempty"
  - "Agent detection populates the model field when available from session data"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Always write model field to checkpoint metadata

## Problem

When checkpoint metadata is serialized, the model field may be omitted entirely if it is empty (due to `omitempty` on the JSON tag). This makes it harder for downstream consumers (dashboards, analytics) to distinguish between "model unknown" and "field missing" — they have to handle both a missing key and an empty value.

## Solution

Remove `omitempty` from the `Model` JSON struct tag in checkpoint metadata types so the field is always present in the serialized JSON (empty string when unknown, never missing). Additionally, ensure all agent detection code paths populate the model field when the information is available from session data.

## Context

- Inspired by entireio/cli#882
- Relevant code: `internal/checkpoint/` (metadata types), `internal/agent/` (detection and model reporting)

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```

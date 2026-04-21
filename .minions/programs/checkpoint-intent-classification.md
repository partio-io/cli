---
id: checkpoint-intent-classification
target_repos:
  - cli
acceptance_criteria:
  - "Post-commit checkpoint creation enriches `metadata.json` with an `intent_tree` field containing classified message nodes"
  - "Each node has a `path` (e.g., `/0/1`), `type` (one of `question`, `task`, `condition`, `none`), and `summary` (first ~80 chars of the message)"
  - "Classification uses simple heuristics (question marks, imperative verbs, conditional keywords) — no external AI calls required"
  - "Existing checkpoints without `intent_tree` continue to work (field is optional)"
  - "Table-driven tests cover classification heuristics and tree construction"
pr_labels:
  - minion
---

# Add intent classification tree to checkpoint metadata

## Description

Enrich checkpoint metadata with a lightweight "intent tree" that classifies each message in the captured session transcript by its epistemic type — whether it seeks information (question), provides information (condition), represents a task/action (task), or is neutral (none). Each message also gets a hierarchical path encoding its position in the conversation flow.

This structured map of the conversation's intent enables future commands (like `partio rewind` or a hypothetical `partio explain`) to navigate long sessions efficiently by their logical structure rather than scanning raw transcripts.

## Motivation

Raw session transcripts capture everything that was said, but not the logical structure of the conversation. For long sessions with many turns, it's difficult to find where key decisions were made or where the agent pivoted approach. An intent tree provides a compact, queryable index into the session.

## Proposed structure

```json
{
  "intent_tree": [
    {"path": "/0", "type": "question", "summary": "How should we handle the race condition in..."},
    {"path": "/0/0", "type": "condition", "summary": "The mutex is held across the entire..."},
    {"path": "/0/0/0", "type": "task", "summary": "Refactor to use channel-based synchronization"},
    {"path": "/0/1", "type": "condition", "summary": "Alternative: use sync.Once for the init path"}
  ]
}
```

## Implementation hints

- Add `internal/checkpoint/intent.go` for the classification logic
- Classification should use simple keyword heuristics (no AI/LLM calls): question marks and interrogative words for `question`, imperative verbs and action phrases for `task`, conditional/explanatory keywords for `condition`
- Build the tree by walking the parsed JSONL messages and tracking conversation depth based on role alternation and topic shifts
- Write the `intent_tree` into `metadata.json` during checkpoint creation in `internal/checkpoint/store.go`
- Make the field optional so existing checkpoints remain valid

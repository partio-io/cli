---
id: conversation-structure-annotations
target_repos:
  - cli
acceptance_criteria:
  - "Post-commit hook generates a conversation_structure field in checkpoint metadata"
  - "Each transcript turn is annotated with a logical type (question, decision, task, context)"
  - "Annotations include a path-style address for navigating the conversation tree"
  - "Annotation generation is best-effort and does not block commit on failure"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add conversation structure annotations to checkpoint metadata

Enrich checkpoint metadata with lightweight structural annotations that classify each transcript turn by its epistemic role (question, decision, task, context) and assign a path address for navigating the conversation's logical flow.

## Why

Raw transcripts show what was said but not the logical structure. When reviewing or resuming a session, it's difficult to programmatically identify where decisions were made, hypotheses formed, or plans laid out. Structural annotations make checkpoints queryable by intent — enabling future commands like `partio explain` to reconstruct reasoning chains without reprocessing the full transcript.

## What to implement

1. After parsing the JSONL transcript in post-commit, analyze each turn and assign:
   - A `type` field: one of `question`, `decision`, `task`, `context`, `none`
   - A `path` field: a hierarchical address (e.g., `/0`, `/0/1`) encoding the turn's position in the conversation's branching structure
2. Store annotations in checkpoint metadata as a `conversation_structure` array alongside the existing transcript data.
3. Use simple heuristics for classification (e.g., turns containing `?` or interrogative patterns → question, turns with tool calls → task). Keep it lightweight — no external API calls.
4. Make this opt-in via a `conversation_structure` config option (default: `false`) since it's experimental.

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```

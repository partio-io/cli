---
id: annotate-reasoning-structure
target_repos:
  - cli
acceptance_criteria:
  - JSONL parser extracts and classifies reasoning blocks from Claude Code transcripts
  - Checkpoint metadata includes a reasoning_nodes field with typed entries
  - Supported node types include at minimum hypothesis, decision, finding, and error-recovery
  - Each reasoning node references the transcript position (line range or message index)
  - partio rewind output displays reasoning node summaries when available
  - Reasoning annotation is optional and does not block checkpoint creation if parsing fails
pr_labels:
  - minion
---

# Add structured reasoning annotations to checkpoint metadata

Enrich checkpoint metadata with classified reasoning nodes extracted from agent transcripts, enabling structured understanding of the agent's decision-making process.

## Motivation

Inspired by entireio/cli#994, which proposed augmenting checkpoint metadata with epistemic node types and path addressing. Partio checkpoints currently store raw transcripts, but understanding *how* the agent reasoned (hypothesized, decided, discovered, recovered from errors) requires reading the full transcript. Structured annotations would make checkpoint data far more useful for session replay, review, and analysis.

## Implementation notes

- Extend the JSONL parser in `internal/agent/claude/` to identify reasoning patterns:
  - **hypothesis**: agent states an assumption or theory about the code
  - **decision**: agent makes a choice between alternatives
  - **finding**: agent discovers something unexpected
  - **error-recovery**: agent encounters a failure and adjusts approach
- Store annotations in checkpoint metadata as a `reasoning_nodes` array, each with `type`, `summary`, and `position` (message index range)
- Pattern matching should be heuristic-based on Claude Code transcript structure (thinking blocks, tool result analysis, retry patterns)
- Annotation failures should log warnings but never block checkpoint creation (consistent with hook error resilience pattern)

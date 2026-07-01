---
id: count-steps-by-prompts
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint metadata includes a step_count field based on user prompt count from the JSONL transcript
  - Sessions with reasoning-only turns (no file changes) report a non-zero step count
  - Existing checkpoints without the field remain readable (backward compatible)
  - Unit tests verify prompt counting across different transcript shapes
pr_labels:
  - minion
---

# Count checkpoint steps by prompt count instead of file-modifying turns

Update checkpoint metadata to count session "steps" based on user prompts (conversation turns) rather than file-modifying turns. Counting only turns that modify files underreports session activity — sessions with extensive reasoning, research, or planning show zero steps even though significant work occurred.

## Implementation Notes

- Parse the JSONL transcript during post-commit checkpoint creation
- Count each user prompt (human turn) as one step, regardless of whether it produced file changes
- Store the prompt-based step count in checkpoint metadata as `step_count`
- Ensure backward compatibility: checkpoints without the field remain readable

## Context

- `internal/checkpoint/` — checkpoint domain type and metadata
- `internal/agent/claude/parse_jsonl.go` — JSONL transcript parsing
- `internal/hooks/` — post-commit hook where checkpoints are created

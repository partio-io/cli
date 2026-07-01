---
id: agent-checkpoint-contributor-interface
target_repos:
  - cli
acceptance_criteria:
  - A CheckpointContributor interface is defined that agents implement to export session-specific data into checkpoint metadata
  - A CheckpointRestorer interface is defined that agents implement to restore session state from checkpoint metadata
  - The existing Claude Code agent implements both interfaces
  - Checkpoint metadata carries agent-specific key/value data through the save/restore lifecycle
  - The rewind command uses CheckpointRestorer to restore agent-specific session state
pr_labels:
  - minion
  - enhancement
---

# Define formal CheckpointContributor/CheckpointRestorer interfaces for multi-agent support

## Problem

Partio currently has a `Detector` interface for identifying which agent is running, but no formal interface for how agents contribute session data to checkpoints or restore from them. The session capture logic is tightly coupled to Claude Code's JSONL format. As more agents are supported, each will have different session formats, resume mechanisms, and metadata needs.

## Desired Behavior

Define two interfaces in the agent package:

```go
type CheckpointContributor interface {
    // ContributeToCheckpoint returns agent-specific metadata to store in the checkpoint
    ContributeToCheckpoint(sessionDir string) (map[string][]byte, error)
}

type CheckpointRestorer interface {
    // RestoreFromCheckpoint takes stored metadata and restores agent session state
    RestoreFromCheckpoint(metadata map[string][]byte, targetDir string) error
    // ResumeHint returns a user-facing message for how to resume the session
    ResumeHint(metadata map[string][]byte) string
}
```

The existing Claude Code implementation would contribute its JSONL transcript and return a resume hint pointing to the session ID. Future agents (Cursor, Codex, etc.) would implement the same interfaces with their own formats.

The checkpoint creation pipeline would call `ContributeToCheckpoint` and store the result. The `rewind` command would call `RestoreFromCheckpoint` and display `ResumeHint`.

## Context Hints

- `internal/agent/detector.go` — existing Detector interface
- `internal/agent/claude/` — Claude Code implementation
- `internal/checkpoint/` — checkpoint creation and storage
- `cmd/partio/rewind.go` — rewind command

## Source

Inspired by entireio/cli#1007 — Cursor agent-lifecycle export/import via CheckpointContributor + CheckpointRestorer interfaces.

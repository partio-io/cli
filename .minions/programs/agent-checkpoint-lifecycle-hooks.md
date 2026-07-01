---
id: agent-checkpoint-lifecycle-hooks
target_repos:
  - cli
acceptance_criteria:
  - a CheckpointContributor interface exists in agent/ that agents can implement to add custom artifacts to checkpoints
  - a CheckpointRestorer interface exists in agent/ that agents can implement to restore artifacts during rewind/resume
  - the Claude Code agent implements CheckpointContributor to bundle session-specific files into checkpoint trees
  - the checkpoint save pipeline calls CheckpointContributor before writing the tree
  - partio rewind calls CheckpointRestorer to extract agent artifacts when restoring a checkpoint
  - agents that don't implement the interfaces are unaffected (backward compatible)
pr_labels:
  - minion
---

# Add checkpoint lifecycle hooks for agent artifact export/import

## Problem

Partio's current agent interface (`agent.Detector`) only handles detection — determining whether an agent is running. Agents have no way to participate in the checkpoint lifecycle: they cannot contribute session-specific artifacts when a checkpoint is saved, nor restore them when a checkpoint is rewound or resumed.

This means agent-specific files (e.g., conversation state, tool call logs, IDE-specific session data) that live outside the JSONL transcript are not captured in checkpoints, reducing the fidelity of checkpoint restore operations.

## Desired behavior

Extend the agent interface with two optional lifecycle hooks:

### CheckpointContributor

Called during checkpoint save. The agent returns a map of relative paths to file contents that should be included in the checkpoint tree alongside the existing transcript and metadata.

```go
type CheckpointContributor interface {
    ContributeToCheckpoint(sessionDir string) (map[string][]byte, error)
}
```

### CheckpointRestorer

Called during `partio rewind` or resume. The agent receives extracted artifacts and restores them to the appropriate locations for the agent to resume from.

```go
type CheckpointRestorer interface {
    RestoreFromCheckpoint(artifacts map[string][]byte, targetDir string) error
}
```

### Integration

- Checkpoint save pipeline checks if the detected agent implements `CheckpointContributor` and includes returned artifacts in the tree
- `partio rewind` checks if the agent implements `CheckpointRestorer` and calls it with extracted artifacts
- Both interfaces are optional — agents that don't implement them work exactly as before

## Context

- Inspired by entireio/cli PR #1007 which adds `CheckpointContributor` and `CheckpointRestorer` for Cursor agent lifecycle
- Currently only Claude Code is implemented as an agent in Partio, but this prepares the architecture for multi-agent support
- The existing `agent.Detector` interface in `internal/agent/detector.go` would remain unchanged; new interfaces are additive
- Checkpoint trees are written via git plumbing in `internal/checkpoint/`

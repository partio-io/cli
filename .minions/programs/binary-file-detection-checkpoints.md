---
id: binary-file-detection-checkpoints
target_repos:
  - cli
acceptance_criteria:
  - "Binary files are detected and excluded from checkpoint tree objects"
  - "Detection uses git's binary heuristic (NUL byte in first 8KB) or file extension matching"
  - "Checkpoint metadata records which files were skipped as binary"
  - "Existing checkpoints without binary detection remain readable"
  - "make test passes with binary detection tests"
  - "make lint passes"
pr_labels:
  - minion
---

# Detect and skip binary files in checkpoint tree writes

Add binary file detection to prevent large binary blobs from being written into checkpoint tree objects.

## What to implement

1. **Binary detection** — before writing file blobs to the checkpoint tree via git plumbing (hash-object, mktree), check if the file content is binary using git's standard heuristic (presence of NUL bytes in the first 8KB)
2. **Skip and record** — exclude detected binary files from the tree object and record the skipped paths in checkpoint metadata
3. **Configurable allowlist** — optionally allow users to configure file extensions that should always be included or excluded via the config system

## Why this matters

When agent sessions modify or create binary files (images, compiled artifacts, database files), these get written into checkpoint tree objects on the orphan branch. This unnecessarily inflates the git object database and slows push operations. Binary content provides no useful context for understanding the reasoning behind code changes.

## Context hints

- `internal/checkpoint/` — checkpoint storage using git plumbing
- `internal/git/` — git operations
- `internal/config/` — configuration system for allowlist settings

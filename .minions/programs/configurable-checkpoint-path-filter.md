---
id: configurable-checkpoint-path-filter
target_repos:
  - cli
acceptance_criteria:
  - "Users can configure `checkpoint_path_filter` in `.partio/settings.json` with include and exclude glob patterns"
  - "Excluded paths are omitted from checkpoint transcript and prompt data"
  - "Include patterns (when specified) restrict checkpoints to only matching paths"
  - "Default behavior (no filter configured) captures all paths as before"
  - "Filters apply during post-commit checkpoint creation"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add configurable path filtering for checkpoint content

Allow users to configure which file paths are included or excluded from checkpoint data via glob patterns in Partio settings.

## Context

Checkpoint data currently captures all files touched during a session. In large monorepos or repos with generated files, build artifacts, or sensitive directories, users may want to exclude certain paths from checkpoint transcripts and metadata. This is distinct from secret redaction — it controls which file paths appear in checkpoint data at all.

## What to implement

1. Add `checkpoint_path_filter` to the config schema with `include` and `exclude` glob arrays
2. Apply path filtering in the post-commit checkpoint creation flow when determining which files to include
3. Exclude patterns take precedence over include patterns when both match
4. Support standard glob syntax (e.g., `vendor/**`, `*.generated.go`, `build/`)
5. Add config validation in `partio doctor` to warn about invalid glob patterns

## Example config

```json
{
  "checkpoint_path_filter": {
    "exclude": ["vendor/**", "node_modules/**", "*.generated.go"],
    "include": ["src/**", "internal/**"]
  }
}
```

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```

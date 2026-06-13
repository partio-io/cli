---
id: settings-schema-versioning
target_repos:
  - cli
acceptance_criteria:
  - Settings files include a schema version field
  - On load, the config package detects the schema version and applies any necessary migrations
  - Unknown future versions produce a clear error suggesting CLI upgrade
  - Existing settings without a version field are treated as v1 and migrated in-place on next write
  - Migration logic is tested with table-driven tests covering v1→v2 transitions
pr_labels:
  - minion
  - enhancement
---

# Add versioned settings schema with migration support

Partio's layered config system (defaults → global → repo → local → env) currently has no schema versioning. As the configuration surface grows, breaking changes require manual user intervention or brittle heuristics.

## What to implement

1. Add a `version` field to settings JSON (default: `1` for current format)
2. In the config loading path, read the version and dispatch to migration functions if needed
3. Implement a migration registry pattern: `migrations[1→2]`, `migrations[2→3]`, etc.
4. On first load of an unversioned file, treat as v1 and write version field on next save
5. If version is higher than the CLI understands, return an error suggesting upgrade

## Why this matters

As Partio adds features (new agent configs, checkpoint strategies, redaction rules), config shape will evolve. Schema versioning prevents silent misinterpretation of old configs and enables automated migration.

## Source

Inspired by entireio/cli PR #1051 — settings schema v2 (parallel to legacy).

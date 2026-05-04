---
id: external-command-plugin-system
target_repos:
  - cli
acceptance_criteria:
  - Running `partio foo` discovers and executes a `partio-foo` binary on $PATH when no built-in command matches
  - Stdin, stdout, stderr, signals, and exit codes pass through transparently to the external command
  - Built-in commands always take precedence over external commands with the same name
  - `partio help` or `partio --help` lists discovered external commands separately from built-in commands
  - External commands receive remaining args via argv with no special protocol required
pr_labels:
  - minion
---

# Add kubectl-style external command plugin system

## Summary

Allow users to extend Partio's CLI by placing `partio-<name>` executables on their `$PATH`. When a user runs `partio <name>`, and no built-in command matches, Partio discovers and execs the external binary, passing through all remaining arguments, stdio, signals, and exit codes transparently.

## Why

Partio's plugin surface is currently limited to the detector interface for agent integration. Users who want to add custom workflows (e.g., custom report generators, team-specific checkpoint analyzers, or integration bridges to internal tools) must fork the CLI or build entirely separate binaries with no discoverability. A kubectl-style plugin system provides a low-friction extension point that requires no registration, no protocol, and no changes to the core CLI — just a binary on PATH.

## What to implement

1. In the Cobra root command's `RunE` or via a `PersistentPreRunE` fallback, detect when no subcommand matches and search `$PATH` for a `partio-<args[0]>` binary.
2. If found, `syscall.Exec` (or `os/exec` with signal forwarding) the external binary with the remaining args.
3. If not found, fall back to Cobra's default unknown-command error.
4. Add a `partio plugins` or listing in `partio help` that scans `$PATH` for `partio-*` binaries and displays them.
5. Reserve the `partio-agent-*` prefix for future agent protocol use.

## Context hints

- `cmd/partio/` — root command setup
- Cobra's `ValidArgsFunction` and unknown-command handling

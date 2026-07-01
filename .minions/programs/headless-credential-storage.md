---
id: headless-credential-storage
target_repos:
  - cli
acceptance_criteria:
  - "Partio can store and retrieve credentials without a system keyring"
  - "File-based credential storage is used as fallback when no keyring is available"
  - "Stored credentials have restricted file permissions (0600)"
  - "Environment variable `PARTIO_CHECKPOINT_TOKEN` can override stored credentials"
  - "Works in CI/CD, Docker, SSH, and WSL environments"
pr_labels:
  - minion
---

# Support file-based credential storage for headless environments

## Summary

Add a file-based credential storage fallback for environments where a system keyring (e.g., gnome-keyring, macOS Keychain) is not available. This enables Partio to store authentication tokens for checkpoint push/fetch in CI/CD pipelines, Docker containers, SSH sessions, and WSL environments.

## Motivation

When `push_sessions` is enabled, Partio needs to authenticate with the remote to push checkpoint refs. Currently, if the system keyring is unavailable, users must set environment variables for every session. A file-based fallback (stored at `~/.config/partio/credentials`) provides a persistent alternative that works in headless environments without requiring env var configuration on every invocation.

## Behavior

1. On `partio login` or first authenticated push, try system keyring first
2. If keyring is unavailable, fall back to `~/.config/partio/credentials` with 0600 permissions
3. `PARTIO_CHECKPOINT_TOKEN` env var always takes precedence over stored credentials
4. `partio doctor` reports which credential storage backend is in use
5. `partio logout` clears credentials from both keyring and file storage

## Context

- Inspired by `entireio/cli` issue #1036 (non-keyring secret storage for headless/CI)
- `internal/config/` for configuration and credential paths
- `PARTIO_CHECKPOINT_TOKEN` env var already documented in CLAUDE.md

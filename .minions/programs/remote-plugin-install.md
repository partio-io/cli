---
id: remote-plugin-install
target_repos:
  - cli
acceptance_criteria:
  - "`partio plugin install <name|url|path>` installs a plugin binary from a git repository, resolving the newest semver tag via `git ls-remote`"
  - "Installed plugins are placed in a managed directory (e.g., `~/.config/partio/plugins/pkg/<name>/`) with a provenance manifest, while symlinks in a `bin/` directory enable dispatch"
  - "`partio plugin upgrade [name|--all]` upgrades installed plugins to the latest version, with `--pin` to lock a specific tag"
  - "A git-synced plugin index (shallow-cloned, TTL-refreshed) enables bare-name installs via `partio plugin search` and `partio plugin info`"
  - "Plugins can declare dependencies via `partio-plugin.yml` `requires` field; install resolves transitive dependencies with cycle detection"
  - "`partio plugin doctor` checks for missing/outdated dependencies, manifest drift, and dangling symlinks"
  - "Existing kubectl-style PATH dispatch (`partio-<name>` in PATH) continues to work unchanged alongside managed plugins"
  - "No forge-specific REST API clients are added; version listing and metadata use the git protocol only"
pr_labels:
  - minion
---

# Add remote plugin install with index discovery and dependency resolution

Extend the existing kubectl-style external plugin system (#356) with remote installation, a discoverable plugin index, and dependency management. Currently, Partio plugins must be manually placed in PATH as `partio-<name>` binaries. This proposal adds a managed install workflow so users can discover, install, upgrade, and manage plugins without manual binary placement.

## Motivation

As Partio grows, the plugin ecosystem needs to support community-built extensions for new agent types, custom attribution strategies, checkpoint processors, and workflow integrations. The current PATH-based discovery is sufficient for power users but creates friction for adoption. A managed install system lowers the barrier to plugin adoption while preserving the zero-cost dispatch model.

Inspired by entireio/cli PR#1422 which implements a full plugin lifecycle with forge-agnostic git-based version resolution.

## Key Design Points

- **Remote install**: `partio plugin install <name|url|path>` resolves the newest semver tag via `git ls-remote` (forge-agnostic). Release assets are downloaded through a per-host URL convention table (GitHub/Gitea-style, GitLab-style, template escape hatch). Binaries land in a managed `pkg/<name>/` directory with a provenance `manifest.yml`.
- **Index discovery**: A krew-style git-synced index (e.g., `partio-io/plugin-index`) shallow-cloned into the user cache with TTL refresh. Supports `search`, `info`, `browse`, and `index update` subcommands. Bare-name installs resolve through the index; raw URLs require confirmation or `--yes`.
- **Dependencies**: `partio-plugin.yml` `requires` field with name, repo_url, and `min_version`. Install-time transitive dependency planning with cycle detection. Remove guard for depended-on plugins.
- **Upgrade**: `partio plugin upgrade [name|--all]` with `--pin` to hold a tag.
- **Doctor**: `partio plugin doctor` checks missing/outdated deps, manifest drift, dangling symlinks.
- **Backward compatible**: PATH dispatch unchanged, built-ins win, existing `partio-<name>` binaries continue to work.

## Context Hints

- `cmd/partio/` — CLI command registration
- `internal/config/` — plugin settings (index URL, TTL)

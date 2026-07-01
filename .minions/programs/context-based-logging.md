---
id: context-based-logging
target_repos:
  - cli
acceptance_criteria:
  - "Logger is created once in main.go and injected into cobra commands via context.Context"
  - "Package-level logger globals and associated mutexes are removed"
  - "WithSession, WithComponent, and similar enrichment functions return a new context with enriched logger"
  - "All commands, hooks, and helpers access the logger from context rather than package globals"
  - "Existing log output format and levels are preserved"
pr_labels:
  - minion
  - refactor
---

# Inject logger via context, remove package globals

## Description

Replace the package-level `*slog.Logger` global (and any associated mutex) in the logging package with a logger that is created once at startup and flows through `context.Context`. This improves testability, removes global mutable state, and makes it straightforward to add per-request or per-hook log enrichment.

## Implementation Notes

- Create the logger in `main.go` and store it in the cobra root command's context
- Add `log.FromContext(ctx)` helper to retrieve the logger
- Convert `WithSession`/`WithComponent` helpers to accept and return `context.Context`, using `slog.With` to add structured fields
- Update all command handlers and hook implementations to pass context through
- Remove the package-level logger variable and any init/sync mechanisms
- Ensure log level configuration still works (via `PARTIO_LOG_LEVEL` env var)

## Context Hints

- `internal/log/` — current slog-based logging setup
- `cmd/partio/` — CLI command entry points
- `internal/hooks/` — hook implementations that use logging

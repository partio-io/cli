---
id: auth-token-lifecycle-management
target_repos:
  - cli
acceptance_criteria:
  - partio status shows authentication status for checkpoint remote (authenticated, expired, or not configured)
  - partio doctor warns when a stored token is expired or about to expire
  - Token refresh is attempted automatically before checkpoint push/fetch when token is near expiry
  - Clear error message when push/fetch fails due to authentication, with instructions to re-authenticate
pr_labels:
  - minion
---

# Add auth token lifecycle management for checkpoint remote

## Summary

Add token lifecycle awareness (expiry detection, refresh, and clear diagnostics) to the checkpoint push/fetch authentication flow. Currently, when `PARTIO_CHECKPOINT_TOKEN` or a stored credential expires, checkpoint operations fail with opaque git errors rather than actionable auth-specific messages.

## What to implement

1. Add token status checking in `internal/checkpoint/` before push/fetch operations:
   - Parse JWT expiry (if token is a JWT) or test the token with a lightweight API call
   - If expired or about to expire (< 5 min), attempt refresh or surface a clear error

2. Add auth status to `partio status` output — show whether checkpoint remote authentication is configured and valid.

3. Add auth health check to `partio doctor` — warn if token is missing, expired, or the remote is unreachable with current credentials.

4. Improve error messages when git push/fetch to the checkpoint remote fails with 401/403 — detect the HTTP status and suggest `partio login` or token re-configuration.

## Context

- `internal/checkpoint/` — checkpoint storage and push/fetch logic
- `internal/config/` — where token/remote config is read
- `cmd/partio/doctor.go` — health checks
- `cmd/partio/status.go` — status display

## Why

Authentication failures during `git push` in the post-commit hook are confusing — the error surfaces as a git plumbing failure rather than an auth problem. Proactive token lifecycle management prevents silent checkpoint loss. Inspired by entireio/cli PR #1050 (better auth token management from the CLI).

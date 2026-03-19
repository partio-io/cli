# Contributing to Partio

Thanks for your interest in contributing to partio! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/partio-io/cli.git
cd cli
make build
```

Requires Go 1.25+.

## Running Tests

```bash
make test    # Run all tests
make lint    # Run golangci-lint
```

## Project Structure

```
cmd/partio/       CLI commands (one file per command)
internal/
  agent/          Agent detection interface + Claude Code implementation
  attribution/    Code attribution calculation
  checkpoint/     Checkpoint domain type + git plumbing storage
  config/         Layered configuration
  git/            Git operations and hook management
  hooks/          Git hook implementations (pre-commit, post-commit, pre-push)
  session/        Session lifecycle and state persistence
  log/            Logging setup
```

## Making Changes

1. Fork the repo and create a feature branch
2. Make your changes
3. Run tests: `make test`
4. Run linter: `make lint`
5. Open a pull request with a clear description of the change

## Code Style

- **One primary concern per file** — e.g., `find_session_dir.go`, `parse_jsonl.go`
- **Standard library `testing` only** — no external test frameworks
- **Table-driven tests** with `t.TempDir()` for filesystem isolation
- **Minimal dependencies** — currently only `cobra` and `google/uuid`
- **Error resilience in hooks** — hooks should log warnings but never block Git operations on non-critical failures

## Reporting Issues

When reporting a bug, please include the output of:

```bash
partio doctor
partio version
```

Along with:
- Your OS and Go version
- Steps to reproduce the issue

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](../LICENSE).

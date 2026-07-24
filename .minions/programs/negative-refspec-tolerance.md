---
id: negative-refspec-tolerance
target_repos:
  - cli
acceptance_criteria:
  - "`partio enable`, `partio status`, and `partio doctor` succeed in a repo whose `.git/config` contains a negative fetch refspec (`fetch = ^refs/pull/*/head`)"
  - "A test with a repo containing a negative refspec is added and passes"
  - "`make test` passes"
  - "`make lint` passes"
pr_labels:
  - minion
  - bug
---

# Tolerate negative fetch refspecs when opening git repositories

## Background

Inspired by entireio/cli PR#1711 (fixes entireio/cli#778).

Git 2.29+ added support for negative (exclusion) refspecs in `.git/config`, e.g.:

```ini
[remote "origin"]
    fetch = +refs/heads/*:refs/remotes/origin/*
    fetch = ^refs/pull/*/head
```

go-git's refspec parser rejects the `^` prefix with a parse error. Since Partio uses go-git to open repositories (e.g. when reading the git config for repo info), any Partio command in a repo with negative refspecs fails at repository-open time with an unhelpful error.

## What to implement

Filter out or gracefully skip negative refspecs when opening or reading remote configuration:

- Before passing refspecs to go-git, strip lines beginning with `^`
- Or catch the parse error from go-git and retry with negative refspecs removed
- Ensure this applies wherever Partio opens a git repository (`internal/git/`)

## Why this matters

Negative refspecs are common in repos that mirror pull request refs or exclude large ref namespaces. Users in such repos see every Partio command fail with a cryptic repository-open error and cannot use Partio at all.

## Context hints

- `internal/git/` (repo open / config reading)
- Any code that calls go-git `PlainOpen` or reads remote fetch refspecs

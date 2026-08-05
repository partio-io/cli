# 01 — Lint gate for swallowed errors

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

Turn on errcheck's blank-discard rule so that assigning an error to
`_` fails lint everywhere it runs — locally, in the minion's per-build
check gate, and in the audit workflows' own checks — and clean up every
existing violation so the gate starts green.

The motivating instance: a hook resolved agent liveness with
`agentStillRunning, _ := detector.IsRunning()`, silently degrading to
the old buggy behavior whenever the check errors. After this slice,
that line does not compile past lint without either handling the error
or logging it deliberately.

## User stories covered

12, 13, 14.

## Acceptance criteria

- [x] The repo lint config enables errcheck's blank-discard checking,
      and the setting demonstrably applies: a scratch blank-discarded
      error fails `golangci-lint run` locally. (The existing config is
      schema v2 — verify the settings block is actually honored, not
      silently ignored; the current file carries a `linters-settings`
      block whose placement may be inert in v2.)
- [x] Test files (`_test.go`) are exempt via lint config exclusions,
      not via code changes.
- [x] Zero blank-discard violations remain in non-test code
      (36 at PRD time; recount, the number may have drifted).
      (Recounted 42 — fixed all 42.)
- [x] Each fix handles the error meaningfully for its context — at
      minimum a debug log naming the operation; no fix silences by
      renaming or restructuring to dodge the linter.
- [x] The liveness case in the post-commit hook logs its error, per the
      PR #586 review recommendation.
      (The surviving blank-discarded `IsRunning` was in the
      pre-commit hook — `precommit.go:42`; the post-commit instance
      no longer exists. Fixed there, logged at Warn.)
- [x] `go test ./...` and `golangci-lint run` both pass.

## Modules touched

- lint tightening (PRD Implementation Decisions).

## Test prior art

- Table-driven unit tests throughout `internal/*/..._test.go` — any
  fix that changes behavior (not just adds logging) extends the
  nearest existing table.
- The lint gate itself is the test for the rule; prove it with a
  deliberate violation before removing it.

## Out of scope

- Any audit workflow or program (issues 02–04).
- Changing what `IsRunning` means — that is the session-liveness
  proposal (issue 05).

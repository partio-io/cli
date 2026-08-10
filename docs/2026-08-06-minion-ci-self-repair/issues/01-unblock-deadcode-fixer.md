# 01 — Unblock the dead-code fixer

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

The dead-code audit's repair session currently refuses to fix any
finding that involves a capitalized (exported) identifier, even when
the audit has fully proven the code inert. It writes no patch, the
first verdict stands, and the pull request goes red for a human to
resolve. Remove that refusal.

The rule exists to protect an external contract. This module has none:
every Go file lives under `internal/` or `cmd/`, and Go forbids
importing an `internal/` package from outside the module that declares
it. There is no consumer any repair here can break.

Concretely, the fix program that runs after a failed dead-code audit
loses its exported-identifier prohibition and the reasoning paragraph
that justified it. Its other restraints are unchanged and must survive
the edit: repair only what a finding's reasoning proves, decline
findings that need a judgment the reasoning does not settle, make no
opportunistic edits beyond the findings, keep each repair coherent so
the tree still builds, and export nothing at all when the repo's checks
cannot be made green.

Record the new rationale in the program itself — that this module
exposes no importable API, so an exported identifier here is not a
public contract — so a future reader knows the rule was scoped
deliberately rather than dropped by accident.

## User stories covered

PRD user stories 1, 2, 3, 30.

## Acceptance criteria

- [x] The dead-code fix program no longer contains a prohibition on
      changing or removing exported identifiers.
- [x] The program states why the prohibition does not apply here: no
      package in this module is importable from outside it.
- [x] The program's remaining restraints are still present and
      unweakened — proof-bounded repairs, declining under-determined
      findings, no opportunistic edits, coherent repairs, and exporting
      nothing when the checks cannot pass.
- [x] The program still forbids the session from running any git write
      command and still requires the skip marker as its final act.
- [x] A dry run of the dead-code audit fix program generates a prompt
      that contains no exported-identifier prohibition.
- [x] Post-merge verification: re-running the dead-code audit on PR
      #622 completes with no surviving findings, and the inert `Model`
      field on the session type is removed by the pipeline with no
      human edit.

## Modules touched

The dead-code fix program from the PRD's Implementation Decisions
("Dead-code fix program (modified)"). No Go packages change in this
slice.

## Test prior art

None applicable — this slice changes a prompt program, which is
model-driven and not unit-testable. The PRD names this gap explicitly
in Testing Decisions. The program's contract is enforced downstream
instead: a session that produces a bad patch fails the apply step's
`make lint && make test` and pushes nothing.

Verification for this slice is the dry-run prompt inspection and the
post-merge audit re-run named in the acceptance criteria. Dry-run
character counts are a useful tell that a section survived the edit —
sections outside the program's agent prose are silently dropped by the
runtime.

The `dry_run` input on the Deadcode Audit workflow does not exercise
this program. That path runs `deadcode-audit.md` only, and the `Fix
round` step waits for a verdict status that a dry run never writes.
Prove the fix program directly, with the version CI pins:

```
GOBIN=/tmp/mbin go install github.com/partio-io/minions/cmd/minions@v0.0.13
MINION_AUDIT_DIR=/tmp/.minion-audit GH_TOKEN="$(gh auth token)" \
  /tmp/mbin/minions run .minions/programs/deadcode-audit-fix.md \
  --pr partio-io/cli#622 --dry-run
```

The run prints a per-component character count. Read
`agent_instructions:deadcode-audit-fix`: it must stay near 3044 chars
and the component count must stay 5. A sharp drop means the rules
section was dropped, not edited.

## Out of scope

- Round counting. The repair still gets exactly one attempt; slice
  [02](./02-count-repair-rounds.md) changes that.
- Any change to what the dead-code audit itself reports. Only the
  fixer's permissions change.
- The e2e-need audit and its fix behavior — it has no fix round and is
  not gaining one.
- Re-scoping the rule to a future importable package. Noted in the
  PRD's Further Notes as the thing to do if one ever appears.

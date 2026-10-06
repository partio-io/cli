# 03 — Review verdict and dry-run gate

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [01 — Prefactor: one GitHub client for the minion tools](./01-prefactor-one-github-client.md)

## What to build

The deterministic half of the review, in dry-run form. A new Go
package, `review`, and a new command, `minion-review`, with one
subcommand in this slice: `gate`.

The review program (slice 04) writes one verdict file per issue. The
gate reads it, checks it, and in this slice only records it: it writes
one row to a tracking issue and changes nothing on the reviewed issue.
This follows the audit-gate pattern: a program writes a verdict file,
and tested Go acts on it and fails closed.

**The verdict file** (JSON) carries:

- `issue` — the issue number; it must equal `--issue`
- `outcome` — `keep`, `rewrite` or `close`
- `close_reason` — for a close: `built`, `false-premise`,
  `does-not-apply`, `duplicate` or `could-not-verify`
- `duplicate_of` — for a duplicate close: the issue that stays
- `premise` — `verdict` (`holds`, `fails`, `unresolved` or
  `no-claims`) and `claims`, each with `claim`, `evidence`, `verdict`
  (`holds`, `fails` or `unresolved`) and `excerpt`
- `fit` — `applies` (bool) and `reason`
- `built` — `built` (bool) and `evidence`
- `duplicates` — each candidate considered: `issue`, `same` (bool),
  `why`
- `rewrite` — for a rewrite: `title`, `body`, `changes` (list)

**The gate fails closed.** "No verdict" covers: a missing file;
malformed JSON; an issue number that differs from `--issue`; an
unknown outcome, reason or claim verdict; a claim without evidence or
excerpt; a close without the evidence for its reason (false premise
needs a failing claim, could-not-verify an unresolved one, built needs
built evidence, does-not-apply needs `applies: false` with a reason,
duplicate needs `duplicate_of` and a matching candidate with
`same: true`); a keep or rewrite that contradicts its own parts (a
failing premise, `applies: false` or `built: true`); and a rewrite
without a title or a body. Slice 07 adds the full rewrite check.

**The tracking issue.** There is one. The gate finds it among open
issues with the label `minion-review-log`, by a marker on the first
line of its body. When it is absent, the gate creates the label and the
issue. Each night has one comment, found by a dated marker and updated
in place. Each reviewed issue adds one row: number, title, verdict or
"no verdict" with its cause, reason, and link. A proposed rewrite goes
in full into a collapsed section under its row. The night is the UTC
date that `--night` names, and its default is today's UTC date. The
sweep workflow passes it once per run, so a run that crosses midnight
writes one comment.

**The command.** `minion-review gate --issue N --verdict <path>
--dry-run [--night YYYY-MM-DD]`. It reads `GITHUB_REPOSITORY`,
`GH_TOKEN` and `GITHUB_API_URL` like the other minion commands. Exit 0
for a valid verdict; exit 1 for no verdict, after it writes the "no
verdict" row; exit 2 for a usage or environment error. Only `--dry-run`
is supported in this slice; without it the command exits 2.

## User stories covered

23, 29, 48, 49, 50 (in part: a malformed verdict is no verdict).

## Acceptance criteria

- [ ] The `review` package defines the verdict contract as a Go type with the JSON field names above.
- [ ] The loader returns "no verdict" for each case listed above, and each case has a test.
- [ ] A close is accepted only with the evidence its reason needs, and a keep or rewrite that contradicts its own parts is rejected.
- [ ] `minion-review gate --dry-run` with a valid verdict writes exactly one row and makes no other request against the reviewed issue: no comment, no label, no edit, no close.
- [ ] The gate finds the tracking issue by label and body marker, never by title or author, and creates the label and the issue once when they are absent.
- [ ] One night's rows share one comment, found by its dated marker and updated in place; a second issue on the same night appends a row.
- [ ] A dry-run row for a rewrite carries the full proposed body in a collapsed section.
- [ ] A missing or malformed verdict writes a "no verdict" row with the cause, and the command exits 1.
- [ ] The gate uses the shared GitHub client from slice 01; tests run it against a fake GitHub server and assert the exact requests.
- [ ] The command is a thin wrapper with no tests of its own, like the other minion commands.
- [ ] `make test` and `make lint` pass.

## Modules touched

- `review` package (new)
- `minion-review` command (new)
- shared GitHub client (issue create, label create, issue list by label)

## Test prior art

- `internal/auditgate/run_test.go` — `fakeGitHub` over `httptest`,
  fail-closed tests for a missing and a malformed verdict, one comment
  upserted in place
- `internal/checksverdict/*_test.go` — verdict conversion tests
- `cmd/minion-audit-gate` — the shape of a thin command over a gate
  package

## Out of scope

- Any change to the reviewed issue: keep and close (slice 06), rewrite
  (slice 07).
- Build mode and its step outputs (slice 09); facts-only mode
  (slice 10).
- `next` and the totals in the tracking issue body (slice 08).
- `dupes` (slice 05) and the review program (slice 04).

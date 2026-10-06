# 06 — Sweep acts on keep and close

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [04 — Review program and dry-run dispatch](./04-review-program-and-dry-run-dispatch.md)

## What to build

The gate stops being dry-run only. Without `--dry-run`, `minion-review
gate` acts on the reviewed issue in the sweep, with the full check.

**Keep.** Add `minion-reviewed`. Remove `minion-approved`,
`minion-failed`, `minion-executing` and `do-not-build` where present:
the March approvals came from automation, and the build labels describe
runs of the old text. Post one evidence comment.

**Close.** Post the evidence comment, add `minion-reviewed`, and close
the issue with GitHub's state reason: `completed` for built, and
`not_planned` for false premise, does not apply, duplicate and could
not verify. A duplicate close names and links the issue that stays.
`minion-reviewed` on a closed issue matters: when the operator reopens
it, the sweep does not select it again.

**The evidence comment** starts with the marker `<!-- partio:review:v1
-->`. It states the outcome and the reason, every claim with its
evidence, verdict and excerpt, and the fit, built and duplicate
decisions. When the issue has an open pull request from an older build
(head branch `minion/implement-implement-<N>`), the comment names it.
The gate never changes a pull request.

**The label** `minion-reviewed` is new. The gate creates it once when
the repository lacks it. Removal of a label that the issue does not
carry is not an error.

**A rewrite verdict** is not supported yet. In this slice the gate
treats it as no verdict and changes nothing. Slice 07 adds it.

Each action also writes its row to the tracking issue, as the dry run
does. The sweep workflow passes its `dry_run` input through, so a
dispatch with `dry_run: false` acts.

## User stories covered

4, 5, 6, 7 (link), 8, 9, 10, 11, 12 (sweep side), 20, 21, 46 (keep
comment), 47.

## Acceptance criteria

- [x] Keep adds `minion-reviewed`, removes `minion-approved`, `minion-failed`, `minion-executing` and `do-not-build` where present, and posts one evidence comment.
- [x] Close posts the evidence comment, adds `minion-reviewed`, and closes the issue.
- [x] The close uses state reason `completed` for built and `not_planned` for every other reason.
- [x] A duplicate close names and links the issue that stays.
- [x] The evidence comment starts with the review marker and lists every claim with evidence, verdict and excerpt, plus the fit, built and duplicate decisions.
- [x] An open pull request from an older build of the issue is named in the comment, and the gate makes no request that changes a pull request.
- [x] The gate creates `minion-reviewed` once when the repository lacks it, and the removal of an absent label does not fail the run.
- [x] A rewrite verdict counts as no verdict in this slice.
- [x] Every action writes its row to the tracking issue.
- [x] The sweep workflow passes the dispatch's `dry_run` input to the gate.
- [x] Tests run against a fake GitHub server and assert the exact requests for keep, for each close reason, and for the label creation.
- [x] `make test` and `make lint` pass.

## Modules touched

- `review` package
- `minion-review` command
- sweep workflow
- shared GitHub client (labels, issue close with state reason, pull
  request list by head)

## Test prior art

- `internal/auditgate/run_test.go` — one test per outcome against a
  fake GitHub server, with exact request assertions
- The stage gate's label handling in the premise programs — remove a
  label first, and treat an absent label as no error

## Out of scope

- Rewrites (slice 07).
- `next`, the schedule and the totals (slice 08).
- Build mode (slice 09) and facts-only mode (slice 10).

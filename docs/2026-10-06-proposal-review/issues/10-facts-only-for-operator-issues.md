# 10 — Facts-only check for the operator's issues

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [09 — Review before each build](./09-review-before-each-build.md)

## What to build

The review never closes or rewrites an issue that the operator wrote,
or one that the operator reopened after a review closed it. For those
issues it checks facts only. Go decides this, not the model.

**The mode rule** (in the gate):

- **Facts only** when the issue has no `minion-proposal` label (the
  operator wrote it), or when its timeline shows a `reopened` event
  after the review's close comment (the comment with the
  `<!-- partio:review:v1 -->` marker on a close). A reopen is the
  operator's overrule. The runner posts as the operator, so the gate
  cannot tell the actors apart, but the review itself never reopens an
  issue.
- **Full check** otherwise.

**Facts-only behavior.** The gate ignores keep, rewrite and close, and
applies only the premise result of the verdict:

- `holds` or `no-claims`: proceed (`blocked=false`), and post an
  evidence comment, so that a pass also leaves evidence.
- `fails` or `unresolved`: remove `do-not-build` and then add it again,
  so that the label event fires again (the stage gate's rule). Post the
  marked comment, which starts with `<!-- partio:premise-gate:v1 -->`
  (kept for continuity) and lists every claim with its evidence,
  verdict and excerpt. Then stop (`blocked=true`).

The gate never closes such an issue, never edits its title or body, and
changes no label on it other than `do-not-build`. The operator
overrules a facts-only block as today: remove the label and run the
build again, or correct the issue text.

The same mode rule applies in the sweep. The sweep does not select
these issues (no `minion-proposal` label, or `minion-reviewed` already
present), but an explicit dispatch list could name one, and then the
gate must not close or rewrite it either.

**August tests.** The tests that assert "a blocked build never closes
the issue" and "the operator overrules the build by removing the label"
now hold for facts-only issues. Rewrite them to state and test that
case.

## User stories covered

11, 12 (build side), 36, 37, 46 (facts-only pass comment).

## Acceptance criteria

- [x] An issue without `minion-proposal` gets facts-only mode.
- [x] An issue with a `reopened` event after the review's close comment gets facts-only mode, and an issue that the review never closed does not.
- [x] The gate reads the issue timeline through the shared GitHub client.
- [x] In facts-only mode, a keep, rewrite or close verdict causes no close, no title or body edit, and no label change other than `do-not-build`.
- [x] A premise that holds, or has no checkable claim, proceeds with `blocked=false` and an evidence comment.
- [x] A premise that fails or is unresolved removes and then adds `do-not-build`, posts the marked comment, and sets `blocked=true`.
- [x] The mode rule also applies in the sweep: an operator issue in a dispatch list is never closed or rewritten.
- [x] The August tests for "never closes" and "overrule by removing the label" now state and test the facts-only case, and they pass.
- [x] Tests run against a fake GitHub server for each mode trigger and each premise result.
- [x] `make test` and `make lint` pass.

## Modules touched

- `review` package
- shared GitHub client (issue timeline)
- premise repo tests

## Test prior art

- `internal/premise/repo_build_test.go` — the "never closes" and
  "overrule by removing the label" tests that this slice rewrites
- The stage gate's text — the remove-then-add label rule and the
  marked comment
- `internal/auditgate/run_test.go` — fake GitHub with exact request
  assertions

## Out of scope

- Any change to the research program's stop behaviour.
- A separate override label: reopening is the override.

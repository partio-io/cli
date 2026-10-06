# 07 — Sweep acts on rewrites

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [06 — Sweep acts on keep and close](./06-sweep-acts-on-keep-and-close.md)

## What to build

The gate applies a rewrite, after it checks that the new body is a
valid proposal. A body that fails the check never replaces a working
one.

**Rewrite validation.** The new body must have the issue shape that
today's proposer files:

- a `## Premise` section that the premise package accepts (the
  `<!-- partio:premise:v1 -->` marker is present, and every claim
  carries `[evidence: …]`)
- an `## Acceptance Criteria` section with at least one `- [ ]` item
- a final `Proposal id: <id>` line
- the original source reference, when the old body had one (20 open
  proposals state no source, and 3 come from the e2e audit)
- no `<!-- program: … -->` pointer to a proposal file

The new title must not be empty. A failed check is "no verdict": the
issue does not change, and the tracking row names the check that
failed.

**A valid rewrite** edits the title and the body in place. GitHub's
edit history keeps the old text. The gate applies the keep label rules
from slice 06 and posts the evidence comment, which also lists what
changed (the verdict's `rewrite.changes`).

In dry run, the row carries the full proposed body (slice 03) and,
for an invalid rewrite, the failed check.

## User stories covered

13, 15, 16, 17, 18, 19, 50.

## Acceptance criteria

- [ ] A rewrite whose body has no premise block that the premise package accepts is no verdict, and the issue does not change.
- [ ] A rewrite without an acceptance-criteria checklist item, without a final proposal id line, or with an empty title is no verdict.
- [ ] A rewrite that drops the source reference of the old body is no verdict, and an old body without a source needs none.
- [ ] A rewrite that still carries a pointer to a proposal file is no verdict.
- [ ] A valid rewrite edits the title and the body in place and applies the keep label rules.
- [ ] The evidence comment of a rewrite lists what changed.
- [ ] A tracking row for an invalid rewrite names the check that failed, in dry run and in a real run.
- [ ] Tests run against a fake GitHub server and cover each failed check and the valid case.
- [ ] `make test` and `make lint` pass.

## Modules touched

- `review` package (it uses the premise package's parser; the premise
  package does not change)
- shared GitHub client (issue title and body edit)

## Test prior art

- `internal/premise/premise_test.go` — `Parse` accepts and rejects
  premise blocks
- `internal/premise/testdata/` — fixture issue bodies
- `internal/auditgate/run_test.go` — fake GitHub with exact request
  assertions

## Out of scope

- The build-mode `changed` output (slice 09).
- Facts-only mode, where a rewrite never applies (slice 10).

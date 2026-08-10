# 07 — Stop the gates colliding

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [06 — Repair failing lint and tests](./06-repair-failing-checks.md)

## What to build

Two hardening changes that only matter once a third gate exists and two
of them can repair.

**A per-pull-request lock across the gates.** The dead-code gate and
the lint/test gate can both fail on the same pull request, and both can
push a repair to the same branch. Nothing today stops them doing it at
the same time, which loses one repair or produces a confusing branch
state. All three gates share a concurrency group keyed on the pull
request number, so at most one repair pushes to a branch at a time.

The existing arrangement has a subtlety that must survive: concurrency
lives on the auditing job, not on the workflow. A repair's own push
triggers a follow-up run, and a workflow-level group would let that run
cancel the in-flight job doing the repairing. A job that skips without
auditing never enters the group, so only runs that will actually do
work supersede each other.

**A single retry for a crashed e2e session.** The e2e-need audit never
fails a pull request on findings — they become proposal issues. Its
only red state is infrastructure: a session that crashes, or one that
finishes without writing a verdict. That is not a code defect and
should not read as one. Retry the audit session once; if the second
attempt also produces nothing readable, the check goes red rather than
being treated as a pass.

## User stories covered

PRD user stories 20, 21, 22, 24.

## Acceptance criteria

- [ ] All three gates share a concurrency group keyed on the pull
      request number.
- [ ] The group is scoped to the job that audits, not to the workflow,
      so a repair's follow-up run cannot cancel the job performing the
      repair.
- [ ] A job that skips without auditing does not enter the group.
- [ ] Two gates failing on the same pull request cannot push repairs
      concurrently.
- [ ] A crashed e2e audit session, or one that writes no verdict, is
      retried once.
- [ ] A second failed attempt leaves the check red; an infrastructure
      failure is never reported as a pass.
- [ ] The e2e audit's judgment is otherwise unchanged: findings still
      become proposal issues and still never fail a pull request.

## Modules touched

All three gate workflows from the PRD's Implementation Decisions
("Workflows"). No Go package changes.

## Test prior art

None applicable — this slice changes workflow configuration, which the
PRD names in Testing Decisions as deliberately not unit-tested. Its
correctness is established by live runs: a pull request failing two
gates at once should show serialized repairs, and the e2e retry is
observable in the run log.

## Out of scope

- Any change to what the e2e audit judges or how it words findings.
- Giving the e2e audit a repair round. It does not block pull requests,
  so it needs none.
- Cross-pull-request serialization. Gates on different pull requests may
  still run concurrently; the runner's own capacity is the only limit
  there.
- Retrying the dead-code or lint/test sessions on a crash. Those already
  fail closed into the verdict gate, and their repair budget covers
  retry.

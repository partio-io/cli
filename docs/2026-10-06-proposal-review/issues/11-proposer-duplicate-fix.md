# 11 — Proposer duplicate fix

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [05 — Duplicate search](./05-duplicate-search.md), [06 — Sweep acts on keep and close](./06-sweep-acts-on-keep-and-close.md)

## What to build

The proposer stops filing an idea a second time, including an idea
that the review closed.

Today the propose program's step 4 checks for an existing proposal
with `gh issue list --repo <this-repo> --label minion-proposal --search
"<feature-id>" --limit 1`. That search has no `--state` flag, so it
sees open issues only, and it matches a kebab-case id that the model
invents on each run. Confirmed misses: #710 and #736, #714 and #720,
#681 and #688. After the sweep closes hundreds of issues, an open-only
check would let the proposer file those ideas again.

Change the propose program:

- Check each idea with `go run ./cmd/minion-review dupes --source
  <source item> --title <title>` (slice 05). It searches open and
  closed proposals by source item and title.
- When a candidate is the same idea, file no issue. Report the idea in
  the run summary as a fourth count, **duplicate**, apart from filed,
  dropped and skipped, and name the issue it matched. A duplicate is
  not a rejection: the matched issue is its record, so it gets no
  rejection-log entry.
- File each new issue with `minion-reviewed` as well as
  `minion-proposal`. The proposer applied today's bar, so the sweep
  must not review the issue again.

Change the propose workflow: before the program runs, make sure the
`minion-reviewed` label exists (`gh label create … --force`). The
proposer can run before any gate run creates it, and `gh issue create
--label` fails on a missing label.

Program text rules: keep instructions as prose in the agent's section
with bold labels. The runtime drops `## Steps` sections and `####`
headings.

## User stories covered

43, 44, 45, 53.

## Acceptance criteria

- [ ] The propose program checks for duplicates with `minion-review dupes`, by the idea's source item and title, across open and closed proposals.
- [ ] It no longer searches open issues by the generated id.
- [ ] It files no issue when a candidate is the same idea.
- [ ] Its summary counts duplicates apart from filed, dropped and skipped items, and names the issue that each duplicate matched.
- [ ] It adds `minion-reviewed` to each issue it files.
- [ ] The propose workflow makes sure the `minion-reviewed` label exists before the program runs.
- [ ] The propose program's instructions stay in its agents section, and the program-shape test passes.
- [ ] The propose repo tests pin the duplicate command, the search of closed issues, the separate duplicate count and the label.
- [ ] `make test` and `make lint` pass.

## Modules touched

- propose program
- propose workflow
- premise repo tests

## Test prior art

- `internal/premise/repo_propose_test.go` — tests that pin what the
  propose program tells the model, including the separation of skipped
  and dropped items
- `internal/programshape/*_test.go` — the shape test

## Out of scope

- The proposer's stuck source cursor (#728 has been open since
  2026-09-26).
- A reader for the rejection log.

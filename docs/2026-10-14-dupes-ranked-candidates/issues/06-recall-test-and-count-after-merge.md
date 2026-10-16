# 06 — Recall test and the count after the merge

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [03 — Code names in the score](./03-code-names-in-the-score.md)

## What to build

Two measures of recall: a test that keeps the found copies found, and
a command that counts the copies that get through after the merge.

**The recall fixture** (new test data in `internal/review/testdata/`).
It comes from the analysis of 2026-10-13, in
`/home/arvos/Code/workspaces/workspace-minion-reboot/review-sweep-tools/`:

- `dupes-analysis/corpus.json`: the 640 proposals of 2026-10-13, in
  the form of `gh issue list --json` (`number`, `title`, `state`,
  `stateReason`, `body`, `createdAt`; states such as `OPEN` and
  `NOT_PLANNED`). Keep the number, title, state, state reason and
  body, gzipped (about 0.5 MB). The fake server serves them in the
  REST form (`open`, `not_planned`), with the `minion-proposal` label.
- The pre-merge bodies of the kept issues that the analysis used:
  `cache/<kept>.body.before-<closed>.md`, for each pair that has one.
  21 pairs have one: 20 hand-found pairs, and the tool-found pair
  704/763. The other files in `cache/` (`before-edit`, `before-reopen`,
  `392.body.before-merge402.md`) were not used, so they stay out.
- `dupes-analysis/pairs-github-dup.tsv` (the 42 hand-found pairs:
  closed issue, then kept issue) and `dupes-analysis/pairs-tool.txt`
  (the 78 tool-found pairs).

**The recall test** goes through the duplicates operation and the fake
server. For each pair, it serves the corpus, with the pre-merge body of
the kept issue when the analysis used one, and runs an issue-number
query for the closed issue. A pair is found when the output (strong or
ranked) holds the kept issue. The test asserts two floors, as counts:
one for the hand-found pairs and one for the tool-found pairs. Each
floor is the count of the first run. On a failure, the test prints
both counts and both floors.

A model of the search (`dupesig`, changed to put the strong matches
first and then the top 10 ranked, with ties to the older issue) found
**30 of 42** hand-found pairs (13 of them by a strong match) and **77
of 78** tool-found pairs (71 by a strong match). The strong matches are
high because the kept issues now hold facets of their copies. A first
run far below these counts is a defect in the search, not a floor.

The model took about 90 ms for each query, so 120 queries take about
11 s in series. Run the pairs as parallel subtests. The test runs in
the default `go test ./...`, with no build tag and no `-short` skip: a
test that does not run guards nothing.

**The count after the merge** (user story 32). Write
`docs/2026-10-14-dupes-ranked-candidates/after-merge.md`, with the
command that lists the proposals filed after the merge and closed by
hand as copies. A hand close of a copy has the state reason
`duplicate`, or a comment that starts with "Duplicate of #": #758 was
closed as `completed`, with the comment "Duplicate of #710 …". The
review gate closes with `not_planned` and a comment that starts with
`<!-- partio:review:v1 -->`, so the command does not count its closes.
This draft was checked on 2026-10-14:

```bash
gh issue list --repo partio-io/cli --label minion-proposal --state closed \
  --search "created:2026-10-07..2026-10-13" --limit 300 --json number,stateReason,comments \
  --jq '[.[] | select(.stateReason == "DUPLICATE" or any(.comments[]; .body | test("^Duplicate of #"))) | .number] | sort'
```

It printed `[757,758,768,771]`: the four copies after #750 that the
PRD names. In the file, the search is `created:>=<merge date>`. The
file gives the command, how to read its output, and the output of the
check window above on the day that the slice runs.

`/code-publish-pr` has no place for this command. So the report of
this slice puts it on the end list: `Ask: put the command of
after-merge.md into the PR body.`

## User stories covered

27, 28, 32.

## Acceptance criteria

- [ ] `internal/review/testdata/` holds the gzipped corpus of 2026-10-13, the 21 pre-merge bodies that the analysis used, and the two pair lists (42 and 78 pairs).
- [ ] The recall test serves the fixture through the fake server, and runs each of the 120 pairs as an issue-number query for the closed issue, with the pre-merge body of the kept issue when one exists.
- [ ] A pair counts as found when the output, strong or ranked, holds the kept issue.
- [ ] The test asserts two floors, written as counts, one for each pair list. Each floor is the count of the first run, and that count is near the model (30 of 42, 77 of 78) or above it.
- [ ] On a failure, the test prints the found count and the floor of each pair list.
- [ ] The recall test runs in the default `go test ./...`, with no build tag and no `-short` skip.
- [ ] `after-merge.md` gives the count command, with `created:>=<merge date>`, and says how to read its output.
- [ ] Live check, read-only: with the window `created:2026-10-07..2026-10-13`, the command prints #757, #758, #768 and #771. The report names any other number as a copy that was closed by hand after 2026-10-14.
- [ ] The end list of the slice report asks for the command in the PR body.
- [ ] `make test` and `make lint` pass.

## Modules touched

- Recall fixture (new files in `internal/review/testdata/`)
- The recall test (a new test file in `internal/review/`)
- `docs/2026-10-14-dupes-ranked-candidates/after-merge.md` (new)

## Test prior art

- `internal/review/dupes_test.go`: `newFakeBacklog` (the paged list endpoint) and `TestFindDupesFindsConfirmedBacklogPairs`, which reads real bodies from `testdata/`.
- `review-sweep-tools/dupes-analysis/dupesig/main.go` (in the workspace, outside the repo): `readPairs` and the lookup of the pre-merge bodies.

## Out of scope

- The full list of all proposal titles, for a judgment of meaning. It comes later, only when the count after the merge shows that copies still get through (PRD).
- The 12 hand-found pairs that the model does not find. They share only common words or files, and only a judgment of meaning finds them.
- The PR body itself: `/code-publish-pr` writes it.

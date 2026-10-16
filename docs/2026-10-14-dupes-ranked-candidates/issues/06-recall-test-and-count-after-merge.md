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

- [x] `internal/review/testdata/` holds the gzipped corpus of 2026-10-13, the 21 pre-merge bodies that the analysis used, and the two pair lists (42 and 78 pairs).
- [x] The recall test serves the fixture through the fake server, and runs each of the 120 pairs as an issue-number query for the closed issue, with the pre-merge body of the kept issue when one exists.
- [x] A pair counts as found when the output, strong or ranked, holds the kept issue.
- [x] The test asserts two floors, written as counts, one for each pair list. Each floor is the count of the first run, and that count is near the model (30 of 42, 77 of 78) or above it.
- [x] On a failure, the test prints the found count and the floor of each pair list.
- [x] The recall test runs in the default `go test ./...`, with no build tag and no `-short` skip.
- [x] `after-merge.md` gives the count command, with `created:>=<merge date>`, and says how to read its output.
- [x] Live check, read-only: with the window `created:2026-10-07..2026-10-13`, the command prints #757, #758, #768 and #771. The report names any other number as a copy that was closed by hand after 2026-10-14.
- [x] The end list of the slice report asks for the command in the PR body.
- [x] `make test` and `make lint` pass.

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

## Evidence

- 1 smoke: ls internal/review/testdata/recall — corpus.json.gz 472K (640 issues), premerge/ 21 bodies, pairs 42 + 78 lines
- 2 green: TestFindDupesRecallOfTheKnownCopies — recall: hand-found 30 of 42, tool-found 77 of 78
- 3 green: TestFindDupesRecallOfTheKnownCopies — recall: hand-found 30 of 42, tool-found 77 of 78
- 4 red: TestFindDupesRecallOfTheKnownCopies — recall below the floor: hand-found 30 of 42 (floor 31), tool-found 77 of 78 (floor 78)
- 4 green: TestFindDupesRecallOfTheKnownCopies — ok github.com/partio-io/cli/internal/review 1.931s (floors 30 and 77)
- 5 red: TestFindDupesRecallOfTheKnownCopies — recall below the floor: hand-found 30 of 42 (floor 31), tool-found 77 of 78 (floor 78)
- 5 green: TestFindDupesRecallOfTheKnownCopies — ok github.com/partio-io/cli/internal/review 1.931s
- 6 smoke: go test ./... -run TestFindDupesRecall -v — PASS in default run, no build tag, no testing.Short skip
- 7 smoke: manual read of after-merge.md — command with created:>=$MERGE_DATE and a how-to-read section
- 8 smoke: gh issue list ... created:2026-10-07..2026-10-13 — [757,758,768,771]
- 10 smoke: make test && make lint (after review fixes) — all packages ok; lint 0 issues
- 9 smoke: end list of the slice message — Ask: put the command of after-merge.md into the PR body.

## Report

**What changed**

- `internal/review/testdata/recall/`: the new fixture.
  - `corpus.json.gz` (472 KB): the 640 proposals of 2026-10-13, with number, title, state, state reason and body.
  - `premerge/`: the 21 pre-merge bodies that the analysis used.
  - `pairs-github-dup.tsv` (42 pairs) and `pairs-tool.txt` (78 pairs).
- `internal/review/dupes_recall_test.go`: the new test `TestFindDupesRecallOfTheKnownCopies`. It runs 120 parallel subtests. Each subtest makes a `FindDupes` issue-number query for the closed issue, through `newFakeBacklog`. The test takes about 2 s.
- `docs/2026-10-14-dupes-ranked-candidates/after-merge.md`: the count command with `created:>=$MERGE_DATE`, how to read its output, and the check window.

**Recall.** The first run found 30 of 42 hand-found pairs and 77 of 78 tool-found pairs. These counts are the same as the model. The floors are 30 and 77. A red run with the floors set to 31 and 78 printed: `recall below the floor: hand-found 30 of 42 (floor 31), tool-found 77 of 78 (floor 78)`.

**Live check (read-only).** On 2026-10-16, the window `created:2026-10-07..2026-10-13` printed `[757,758,768,771]`. No other number came back.

**Suite.** `make test`: all packages pass, before and after the review fixes. `make lint`: 0 issues. The first lint run found 2 `errcheck` issues on `defer f.Close()` in the new test, and argos fixed them.

**Evidence records:** red 2, green 4, smoke 6 (criterion 10 has two records), prefactor 0, screenshot 0.

**Review findings** (`code-review medium --fix`):

1. The floor check failed when `-run` selects only some pairs. Fixed: the test counts the pairs that ran and skips the floors on a partial run. Kept, because it is a defect in the test of criterion 4.
2. The corpus was converted again for each of the 120 pairs. Fixed: one shared base, and a clone only for a pair with a pre-merge body.
3. An extra tab replacement before `fmt.Sscan`. Fixed: removed.
4. `after-merge.md` named the state reason `duplicate`, but the command uses `DUPLICATE`. Fixed.
5. Each pair runs against the full snapshot of 2026-10-13, so issues filed after the closed issue compete for the ranked places. Skipped: a change changes what the test measures. Ask item on the end list.
6. Criterion 9 had no evidence record. Fixed: a `smoke` record now names the end-list item.

**Spec check:** (a) none, (b) none, (c) none.

**Marked with `git add -N`:** `internal/review/dupes_recall_test.go`, every file under `internal/review/testdata/recall/`, `docs/2026-10-14-dupes-ranked-candidates/after-merge.md`.

**Check time:** 08:20:35 to 08:33:14, 12 min 39 s.

**Notes.** The fixture is in `testdata/recall/`, a subfolder, so that it is apart from the four earlier body files. `after-merge.md` uses `MERGE_DATE=YYYY-MM-DD`, because the merge date is not known yet.

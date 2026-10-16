# 04 — The proposer checks with the text of its idea

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [03 — Code names in the score](./03-code-names-in-the-score.md)

## What to build

The proposer gives the duplicate check the text of its idea, and it
judges the ranked candidates. Only the program text and its repo tests
change: `.minions/programs/propose.md`, step 4, "Check for an earlier
proposal of the same idea".

The check stays where it is: before the premise, so that a copy costs
no premise work. At that step the agent has the idea from the ingest
step (title, source, description, acceptance criteria, premise), not
the filed body. There is no second check after the premise: the
proposer has no spare turns (`max_turns: 80`, and one of its last 8
runs stopped at the limit).

The new text of the step:

- **The command**, in a fenced block, with the idea text on standard
  input from a quoted heredoc. Its terminator is not `PROPOSAL_BODY`.
  For example:

  ```bash
  go run ./cmd/minion-review dupes --source "<owner/repo#N>" --title "<title>" --body-file - <<'IDEA_TEXT'
  <the description, the acceptance criteria and the premise of the idea, from the ingest step>
  IDEA_TEXT
  ```

  Keep the rules for `--source`: repeat it for each source item, and
  leave it out for a changelog entry.
- **The three kinds of match**: `"match": "source"`, `"match": "title"`
  and `"match": "ranked"` (one of the 10 other proposals with the best
  score), and the fields of a candidate, `score`, `shared` and
  `excerpt` among them.
- **Read the title and the excerpt of each candidate.** Open a
  candidate with `gh issue view <n>` only when it can be the same idea.
- **"The list is the whole check: do not search the issues in another
  way."** stays.
- **A ranked candidate that is the same idea** counts as a strong match
  does: file no issue. The current rules for closed candidates stay (a
  review close counts; `completed` counts only when the tree shows the
  work).
- **A candidate closed as a copy** (`state_reason` `duplicate`, or a
  comment "Duplicate of #N", as on #758) stands for the issue that it
  names. Apply the rules to that issue.
- **When not sure** that a candidate is the same idea, file the issue,
  and add the line `Possible duplicate of #<n>: <reason>` above the
  source line of the body. No idea gets lost, and a reader sees the
  possible copy. The reason names no `owner/repo` item: `SourceRefs`
  reads an item on that line as a source of the issue.

The text stays in the `## Agents` prose of the `### proposer` agent,
with bold labels and no `####` heading: the minions parser drops every
other H2 section and every `####` heading. A static test cannot prove
that the text reaches the agent; only a real dry run can.

The workflows do not change: they read the program from `main` and run
the command with `go run`, so the merge is live at the next proposer
run, with no minions release and no pin change.

## User stories covered

12, 13, 14, 15, 16, 17, 30 (for the propose program), 31 (for the proposer).

## Acceptance criteria

- [x] The duplicate step shows the command in a fenced block. The command passes `--source`, `--title` and `--body-file -`, with the idea text in a quoted heredoc whose terminator is not `PROPOSAL_BODY`.
- [x] The step says what the idea text holds: the description, the acceptance criteria and the premise from the ingest step.
- [x] The step names the three kinds of match (`source`, `title`, `ranked`) and the fields `score`, `shared` and `excerpt`.
- [x] The step tells the agent to read the title and the excerpt of each candidate, and to open with `gh issue view` only a candidate that can be the same idea.
- [x] "The list is the whole check" stays, and the check still comes before the premise check and before `gh issue create`.
- [x] A ranked candidate that is the same idea leads to no issue, as a strong match does.
- [x] A candidate closed as a copy stands for the issue that it names.
- [x] When not sure, the proposer files the issue with the line `Possible duplicate of #<n>: <reason>` above the source line.
- [x] `internal/premise/repo_propose_test.go` asserts the points above. The current assertions pass (the order of the steps, the same-idea bullet, the heredoc on create, "with no escapes"); they change only where the text changes on purpose.
- [x] A real dry run of the propose program prints the new step 4 text in the agent instructions, and the text of steps 5, 6 and 7 after it. See "Dry run" below.
- [x] No file under `.github/` changes, and `make test` and `make lint` pass.

## Dry run

```bash
GOBIN=<scratchpad>/bin go install github.com/partio-io/minions/cmd/minions@v0.0.14
cd /home/arvos/Code/workspaces/workspace-minion-reboot/cli
GH_TOKEN=$(gh auth token) WORKSPACE_ROOT=/home/arvos/Code/workspaces/workspace-minion-reboot \
  <scratchpad>/bin/minions run .minions/programs/propose.md --dry-run --project .minions/project.yaml
```

`<scratchpad>` is the scratchpad of the session. In the output, check
the agent instructions: the new duplicate step, the fenced command
with its heredoc, and the steps after it.

## Modules touched

- The proposer program (`.minions/programs/propose.md`)
- Its repo tests (`internal/premise/repo_propose_test.go`)

## Test prior art

- `internal/premise/repo_propose_test.go`: `TestProposeProgramChecksDuplicatesBySourceAndTitle` reads the first line that holds "minion-review dupes", so that line must be the command line. Also `TestProposeProgramFilesNoDuplicate`, `TestProposeProgramFilesTheBodyAsWritten` (the heredoc check), and the helpers `section`, `lineContaining` and `fencedBlockContaining`.

## Out of scope

- The count of escaped backticks after the proposer files: [05 — Escaped backticks in the whole body](./05-escaped-backticks-in-the-whole-body.md).
- A second duplicate check after the premise, and the turn limit of the proposer (PRD).
- The label `minion-reviewed` at creation, and the schedule of the review sweep (PRD).
- The review program: [02 — The review checks by issue number](./02-review-checks-by-issue-number.md).

## Evidence

- 1 red: TestProposeProgramChecksWithTheIdeaText — propose program shows no duplicate search command in a code block
- 1 green: TestProposeProgramChecksWithTheIdeaText — ok github.com/partio-io/cli/internal/premise
- 2 red: TestProposeProgramChecksWithTheIdeaText — propose program shows no duplicate search command in a code block (the idea-text assertions follow in the same test)
- 2 green: TestProposeProgramChecksWithTheIdeaText — ok github.com/partio-io/cli/internal/premise
- 3 red: TestProposeProgramJudgesTheRankedCandidates — the proposer is never told `"match": "ranked"`, so it cannot tell a ranked candidate
- 4 red: TestProposeProgramJudgesTheRankedCandidates — the proposer is never told to judge a candidate by its title and excerpt first
- 6 red: TestProposeProgramJudgesTheRankedCandidates — the proposer is never told that a ranked candidate can stop the issue
- 7 red: TestProposeProgramJudgesTheRankedCandidates — the proposer is never told that a closed copy stands for the issue that it names
- 8 red: TestProposeProgramJudgesTheRankedCandidates — the proposer is never told how to file an idea when it is not sure of a match
- 3 green: TestProposeProgramJudgesTheRankedCandidates — ok github.com/partio-io/cli/internal/premise
- 4 green: TestProposeProgramJudgesTheRankedCandidates — ok github.com/partio-io/cli/internal/premise
- 6 green: TestProposeProgramJudgesTheRankedCandidates — ok github.com/partio-io/cli/internal/premise
- 7 green: TestProposeProgramJudgesTheRankedCandidates — ok github.com/partio-io/cli/internal/premise
- 8 green: TestProposeProgramJudgesTheRankedCandidates — ok github.com/partio-io/cli/internal/premise
- 5 green: TestProposeProgramChecksDuplicatesBySourceAndTitle — ok github.com/partio-io/cli/internal/premise (order check unchanged: dupes before verify before create)
- 9 green: go test ./internal/premise — ok github.com/partio-io/cli/internal/premise (old assertions unchanged)
- 10 smoke: minions v0.0.14 run propose.md --dry-run — exit 0; agent instructions show the new step 4 (fenced dupes command, IDEA_TEXT heredoc, ranked, copy, Possible duplicate) and steps 5, 6, 7 after it
- 11 green: make test; make lint; git status .github — make test exit 0, golangci-lint 0 issues, no change under .github/

## Report

**What changed**

- `.minions/programs/propose.md`, step 4: the duplicate search is a fenced command with `--source`, `--title` and `--body-file -`, and the idea text (description, acceptance criteria, premise, from the ingest step) in a `'IDEA_TEXT'` heredoc. The text names the three kinds of match, the candidate fields (`score`, `shared`, `excerpt` among them), and tells the agent to read the title and the excerpt first and to open a candidate only when it can be the same idea. A ranked same-idea candidate stops the issue. A candidate closed as a copy stands for the issue that it names. When not sure, the agent files the issue with `Possible duplicate of #<n>: <reason>` above the source line. The `gh issue create` body template now lists that line.
- `internal/premise/repo_propose_test.go`: new `TestProposeProgramChecksWithTheIdeaText` and `TestProposeProgramJudgesTheRankedCandidates`, and a `proposeAgents` helper. The old assertions did not change.

**Checks**

- Steps 1 and 5: `make test` exit 0, `make lint` 0 issues, `go vet ./...` clean. No module outside "Modules touched" was repaired. No file under `.github/` changed.
- Dry run (`minions` v0.0.14, `--dry-run`), run before and after the review fixes: exit 0. The agent instructions show the new step 4, the fenced command with its heredoc, and steps 5, 6 and 7 after it.
- Evidence: 7 `red`, 10 `green`, 1 `smoke`. Criterion 5 had no red run: the current order test already covers it. Criterion 11 is the suite, lint and `.github/` check.

**Review findings** (`code-review medium --fix`)

1. Fixed, kept (criterion 8): the reason rule now also forbids a path and an other `#N`. `sourceToken` reads `internal/review (` as a repo token, so the next `#N` becomes a source item. Checked against `internal/review/source_refs.go`.
2. Fixed, kept (criterion 8): the `gh issue create` body template now lists the possible-duplicate line before the source line.
3. Fixed, kept (criterion 7): the agent opens the issue that a copy names with `gh issue view <n>`. That issue is often not in the list.
4. Fixed, kept (criterion 3): "one of the up to 10 other proposals with the best score above zero". `FindDupes` returns at most `maxRanked` and only scores above 0.
5. Fixed, kept (refactor): six old tests now use `proposeAgents`.
6. Skipped: map order of the failure messages in the new test. It is cosmetic only.

**Spec findings**

- (a) none. (c) none.
- (b) The `proposeAgents` refactor of six tests: kept. It is the refactor step of this slice and changes no assertion.
- (b) The ban on paths and an other `#N` in the reason: kept. It closes the same `SourceRefs` leak that criterion 8 guards against.

**Other**

- `git add -N`: no files. The slice created no file.
- Time of the parallel checks: 08:14:16 to 08:15:55, 1 min 39 s.
- Surprise: none against the PRD. The text claims the reach of the agent only through the dry run. A real proposer run is the first proof that the agent acts on it.

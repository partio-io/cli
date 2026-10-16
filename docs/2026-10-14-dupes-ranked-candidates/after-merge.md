# The count after the merge

This command counts the copies that get through the search after the
merge of the ranked candidates (user story 32). It lists the proposals
that were filed on or after the merge date and that a person closed by
hand as copies.

## The command

Set `MERGE_DATE` to the day of the merge, as `YYYY-MM-DD`:

```bash
MERGE_DATE=YYYY-MM-DD  # the day of the merge
gh issue list --repo partio-io/cli --label minion-proposal --state closed \
  --search "created:>=$MERGE_DATE" --limit 300 --json number,stateReason,comments \
  --jq '[.[] | select(.stateReason == "DUPLICATE" or any(.comments[]; .body | test("^Duplicate of #"))) | .number] | sort'
```

The command only reads. It needs `gh` with read access to
`partio-io/cli`.

## How to read the output

- The output is a JSON list of issue numbers, sorted. Each number is a
  proposal that a person closed as a copy of an older one.
- A hand close of a copy has the state reason `DUPLICATE`, or a comment
  that starts with `Duplicate of #`. For example, #758 was closed as
  `completed`, with the comment "Duplicate of #710 …".
- The review gate closes with `not_planned` and a comment that starts
  with `<!-- partio:review:v1 -->`. The command does not count these
  closes: the gate stopped those copies.
- `[]` means that no copy got through. Each number in the list is a
  copy that the search did not stop. When the list is not empty, the
  full list of the proposal titles, for a judgment of meaning, is the
  next step (PRD).
- The command reads at most 300 closed proposals. When the window holds
  more, make it shorter.

## The check window

The window before the merge, `created:2026-10-07..2026-10-13`, gives a
known output. Use it to check the command:

```bash
gh issue list --repo partio-io/cli --label minion-proposal --state closed \
  --search "created:2026-10-07..2026-10-13" --limit 300 --json number,stateReason,comments \
  --jq '[.[] | select(.stateReason == "DUPLICATE" or any(.comments[]; .body | test("^Duplicate of #"))) | .number] | sort'
```

Output on 2026-10-16:

```
[757,758,768,771]
```

These are the four copies after #750 that the PRD names.

package rejection

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/premise"
)

// proposeProgram is this repository's propose program, relative to the package
// directory that go test runs in. It is the run that drops ideas, so it is the
// run that has to write and commit the log.
const proposeProgram = "../../.minions/programs/propose.md"

// TestRejectionLogReachesTheProposer is the end-to-end path this slice builds.
// A log nothing writes records nothing, and a log the run never commits is gone
// the moment the workspace is thrown away — while the cursor has already moved
// past the idea. So the test walks both halves: entries survive a write and a
// read back, and the propose program stages the log in the same commit that
// advances the cursor.
func TestRejectionLogReachesTheProposer(t *testing.T) {
	root := t.TempDir()

	dropped := Entry{
		Idea:   "walk the tree to attribute a commit",
		Source: "entireio-cli-issues#1877",
		Reason: PremiseFailed,
		Claim: premise.Claim{
			Statement: "the post-commit hook walks the whole tree",
			Evidence:  "internal/hooks/post_commit.go",
		},
		Verdict: premise.Fails,
		Found:   "the hook names changed files with git.DiffNameOnly(commitHash)",
	}
	skipped := Entry{
		Idea:   "add a billing dashboard",
		Source: "entireio-cli-changelog#0.9.1",
		Reason: Irrelevant,
		Note:   "this project has no billing surface",
	}

	for _, e := range []Entry{dropped, skipped} {
		if err := Append(root, e); err != nil {
			t.Fatalf("append %q: %v", e.Idea, err)
		}
	}

	raw, err := os.ReadFile(filepath.Join(root, LogPath))
	if err != nil {
		t.Fatalf("read the log: %v", err)
	}

	got, err := Parse(string(raw))
	if err != nil {
		t.Fatalf("parse the log: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("wrote 2 rejections, read back %d", len(got))
	}
	if got[0].Reason != PremiseFailed || got[1].Reason != Irrelevant {
		t.Errorf("the two kinds of rejection did not survive the round trip: %q then %q",
			got[0].Reason, got[1].Reason)
	}
	if got[0].Claim.Statement != dropped.Claim.Statement || got[0].Found != dropped.Found {
		t.Errorf("the failed claim lost what killed it: %+v", got[0])
	}

	// The other half: the run has to commit what it wrote. The cursor moves
	// past a dropped idea whether or not the log survives, so a log left out
	// of the commit is the silent loss this slice exists to stop.
	src, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read the propose program: %v", err)
	}
	add, ok := lineContaining(string(src), "git add")
	if !ok {
		t.Fatal("the propose program stages nothing")
	}
	if !strings.Contains(add, LogPath) {
		t.Errorf("the run does not stage %s, so a rejection dies with the workspace:\n%s", LogPath, add)
	}
	if !strings.Contains(add, ".minions/sources.yaml") {
		t.Errorf("the log is not staged alongside the cursor it must travel with:\n%s", add)
	}
}

// TestEntryRecordsIdeaSourceAndReason checks the three fields every entry
// carries whatever the reason. An entry that names no idea, or no source, or no
// known reason cannot be judged later, and a log of unjudgeable entries is the
// silence this slice replaced. So a short entry is refused at the write rather
// than appended half-empty for a reader to puzzle over.
func TestEntryRecordsIdeaSourceAndReason(t *testing.T) {
	complete := Entry{
		Idea:    "walk the tree to attribute a commit",
		Source:  "entireio-cli-issues#1877",
		Reason:  PremiseFailed,
		Claim:   premise.Claim{Statement: "the hook walks the tree", Evidence: "internal/hooks/post_commit.go"},
		Verdict: premise.Fails,
		Found:   "the hook calls git.DiffNameOnly(commitHash)",
	}
	without := func(f func(*Entry)) Entry {
		e := complete
		f(&e)
		return e
	}

	tests := []struct {
		name    string
		entry   Entry
		wantErr error
	}{
		{"complete", complete, nil},
		{"no idea", without(func(e *Entry) { e.Idea = "" }), ErrNoIdea},
		{"blank idea", without(func(e *Entry) { e.Idea = "   " }), ErrNoIdea},
		{"no source", without(func(e *Entry) { e.Source = "" }), ErrNoSource},
		{"no reason", without(func(e *Entry) { e.Reason = "" }), ErrNoReason},
		{"reason outside the closed set", without(func(e *Entry) { e.Reason = "meh" }), ErrNoReason},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			err := Append(root, tt.entry)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("append a complete entry: %v", err)
				}
				got := readOne(t, root)
				if got.Idea != tt.entry.Idea {
					t.Errorf("idea: got %q, want %q", got.Idea, tt.entry.Idea)
				}
				if got.Source != tt.entry.Source {
					t.Errorf("source: got %q, want %q", got.Source, tt.entry.Source)
				}
				if got.Reason != tt.entry.Reason {
					t.Errorf("reason: got %q, want %q", got.Reason, tt.entry.Reason)
				}
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("append %s: got %v, want %v", tt.name, err, tt.wantErr)
			}
			if _, err := os.Stat(filepath.Join(root, LogPath)); !os.IsNotExist(err) {
				t.Error("a refused entry still touched the log")
			}
		})
	}
}

// TestPremiseFailedRecordsTheClaimAndWhatKilledIt checks the entry that has to
// carry the most. "The premise failed" is the assertion the premise block
// replaced: it is a verdict with nothing behind it. An entry that names the
// claim, the evidence it named, the verdict and what the evidence actually
// showed can be re-judged later; one that names fewer cannot, so it is refused.
func TestPremiseFailedRecordsTheClaimAndWhatKilledIt(t *testing.T) {
	complete := Entry{
		Idea:    "walk the tree to attribute a commit",
		Source:  "entireio-cli-issues#1877",
		Reason:  PremiseFailed,
		Claim:   premise.Claim{Statement: "the hook walks the tree", Evidence: "internal/hooks/post_commit.go"},
		Verdict: premise.Fails,
		Found:   "the hook calls git.DiffNameOnly(commitHash) and never walks",
	}
	without := func(f func(*Entry)) Entry {
		e := complete
		f(&e)
		return e
	}

	tests := []struct {
		name    string
		entry   Entry
		wantErr error
	}{
		{"no claim", without(func(e *Entry) { e.Claim.Statement = "" }), ErrNoClaim},
		{"claim with no evidence", without(func(e *Entry) { e.Claim.Evidence = "" }), ErrNoClaim},
		{"no verdict", without(func(e *Entry) { e.Verdict = "" }), ErrNoVerdict},
		{"verdict outside the closed set", without(func(e *Entry) { e.Verdict = "maybe" }), ErrNoVerdict},
		// A claim the tree confirmed did not cause this rejection. Recording
		// one says the gate fired on evidence that supports the idea, which is
		// either a mis-recorded entry or a real bug in the gate — either way
		// the log must not be the place it goes quiet.
		{"a verdict that holds", without(func(e *Entry) { e.Verdict = premise.Holds }), ErrNoVerdict},
		{"no finding", without(func(e *Entry) { e.Found = "" }), ErrNoFinding},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if err := Append(root, tt.entry); !errors.Is(err, tt.wantErr) {
				t.Fatalf("append %s: got %v, want %v", tt.name, err, tt.wantErr)
			}
			if _, err := os.Stat(filepath.Join(root, LogPath)); !os.IsNotExist(err) {
				t.Error("an entry that records nothing about the failure still reached the log")
			}
		})
	}

	t.Run("complete", func(t *testing.T) {
		root := t.TempDir()
		if err := Append(root, complete); err != nil {
			t.Fatalf("append a complete entry: %v", err)
		}

		got := readOne(t, root)
		if got.Claim != complete.Claim {
			t.Errorf("claim: got %+v, want %+v", got.Claim, complete.Claim)
		}
		if got.Verdict != complete.Verdict {
			t.Errorf("verdict: got %q, want %q", got.Verdict, complete.Verdict)
		}
		if got.Found != complete.Found {
			t.Errorf("found: got %q, want %q", got.Found, complete.Found)
		}
	})
}

// TestIrrelevantIsDistinguishableFromPremiseFailed checks the difference the
// log exists to preserve. An irrelevant item was never about this project and
// no claim was ever checked; a dropped idea was checked and the tree
// contradicted it. Folded together they are unreadable, and the reader cannot
// tell a bar set too high from a source that has gone quiet. The difference is
// structural, not just a label: an entry that was never checked cannot carry a
// verdict, and one that was checked cannot omit it.
func TestIrrelevantIsDistinguishableFromPremiseFailed(t *testing.T) {
	irrelevant := Entry{
		Idea:   "add a billing dashboard",
		Source: "entireio-cli-changelog#0.9.1",
		Reason: Irrelevant,
		Note:   "this project has no billing surface",
	}
	with := func(f func(*Entry)) Entry {
		e := irrelevant
		f(&e)
		return e
	}

	tests := []struct {
		name    string
		entry   Entry
		wantErr error
	}{
		{"no note", with(func(e *Entry) { e.Note = "" }), ErrNoNote},
		{
			"carries a claim it never checked",
			with(func(e *Entry) { e.Claim = premise.Claim{Statement: "x", Evidence: "y"} }),
			ErrNotChecked,
		},
		{"carries a verdict it never reached", with(func(e *Entry) { e.Verdict = premise.Fails }), ErrNotChecked},
		{"carries a finding it never gathered", with(func(e *Entry) { e.Found = "something" }), ErrNotChecked},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if err := Append(root, tt.entry); !errors.Is(err, tt.wantErr) {
				t.Fatalf("append %s: got %v, want %v", tt.name, err, tt.wantErr)
			}
		})
	}

	t.Run("the two kinds stay apart in one log", func(t *testing.T) {
		root := t.TempDir()
		failed := Entry{
			Idea:    "walk the tree to attribute a commit",
			Source:  "entireio-cli-issues#1877",
			Reason:  PremiseFailed,
			Claim:   premise.Claim{Statement: "the hook walks the tree", Evidence: "internal/hooks/post_commit.go"},
			Verdict: premise.Fails,
			Found:   "the hook calls git.DiffNameOnly(commitHash)",
		}
		for _, e := range []Entry{failed, irrelevant} {
			if err := Append(root, e); err != nil {
				t.Fatalf("append %q: %v", e.Idea, err)
			}
		}

		raw, err := os.ReadFile(filepath.Join(root, LogPath))
		if err != nil {
			t.Fatalf("read the log: %v", err)
		}
		got, err := Parse(string(raw))
		if err != nil {
			t.Fatalf("parse the log: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("wrote 2 entries, read back %d", len(got))
		}
		if got[0].Reason == got[1].Reason {
			t.Fatalf("both entries read back as %q", got[0].Reason)
		}
		if got[1].Verdict != "" || got[1].Claim.Statement != "" || got[1].Found != "" {
			t.Errorf("the irrelevant entry came back carrying a check that never ran: %+v", got[1])
		}
		if got[1].Note != irrelevant.Note {
			t.Errorf("note: got %q, want %q", got[1].Note, irrelevant.Note)
		}
	})

	// The program half. The proposer is the only writer of this log, so a
	// distinction the package enforces and the program never uses is a
	// distinction the log never actually carries.
	t.Run("the proposer records both kinds", func(t *testing.T) {
		raw, err := os.ReadFile(proposeProgram)
		if err != nil {
			t.Fatalf("read the propose program: %v", err)
		}
		agents, ok := section(string(raw), "## Agents")
		if !ok {
			t.Fatal("the propose program has no agents section")
		}

		// The field form, not the bare word: "irrelevant" already appears in
		// the run summary as prose, and a test that matches that passes on a
		// program which records nothing.
		for _, r := range Reasons {
			field := "- reason: `" + string(r) + "`"
			if !strings.Contains(agents, field) {
				t.Errorf("the proposer is never shown %q, so it records no rejection of that kind", field)
			}
		}
	})
}

// TestDryRunWritesAndCommitsNothing checks where the side effects live. A dry
// run renders the program and starts no agent, so anything the agent is told to
// do costs nothing. A write or a commit stated outside the agent's steps is a
// side effect of the program itself, and a dry run would perform it — which
// would put a rejection in the log for a run that never happened.
func TestDryRunWritesAndCommitsNothing(t *testing.T) {
	raw, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read the propose program: %v", err)
	}
	src := string(raw)

	agents, ok := section(src, "## Agents")
	if !ok {
		t.Fatal("the propose program has no agents section")
	}

	outside := strings.Replace(src, agents, "", 1)
	for _, effect := range []string{"git add", "git commit", "git push", LogPath} {
		if !strings.Contains(agents, effect) {
			t.Errorf("%q is not one of the agent's steps", effect)
		}
		if strings.Contains(outside, effect) {
			t.Errorf("%q sits outside the agent's steps, so a dry run performs it", effect)
		}
	}
}

// TestRunThatRejectsNothingStillCommits checks the quiet run. Most runs reject
// nothing, and the run still has to commit the cursor it advanced. The log is
// staged unconditionally, so it has to exist in this repository: `git add` on a
// path that has never been created fails, and a failed stage takes the cursor
// commit down with it — every run, not just the quiet ones.
func TestRunThatRejectsNothingStillCommits(t *testing.T) {
	inRepo := "../../" + LogPath

	raw, err := os.ReadFile(inRepo)
	if err != nil {
		t.Fatalf("the log is staged every run but does not exist in this repository: %v", err)
	}
	if _, err := Parse(string(raw)); err != nil {
		t.Errorf("the log in this repository does not parse: %v", err)
	}

	// A run that appends nothing leaves the file byte-identical, so the stage
	// is a no-op and the commit carries the cursor alone.
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(LogPath)), 0o755); err != nil {
		t.Fatalf("seed the log directory: %v", err)
	}
	seeded := filepath.Join(root, LogPath)
	if err := os.WriteFile(seeded, raw, 0o644); err != nil {
		t.Fatalf("seed the log: %v", err)
	}

	before, err := os.Stat(seeded)
	if err != nil {
		t.Fatalf("stat the seeded log: %v", err)
	}
	got, err := os.ReadFile(seeded)
	if err != nil {
		t.Fatalf("read the seeded log: %v", err)
	}
	if string(got) != string(raw) || before.Size() != int64(len(raw)) {
		t.Error("a run that rejected nothing changed the log")
	}

	// And the program says so, so the agent does not invent a second commit or
	// skip the cursor when it has no rejections to report.
	src, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read the propose program: %v", err)
	}
	agents, ok := section(string(src), "## Agents")
	if !ok {
		t.Fatal("the propose program has no agents section")
	}
	if !strings.Contains(agents, "rejected nothing") {
		t.Error("the proposer is never told what to do when it rejects nothing")
	}
}

// TestEntriesAccumulateAcrossRuns checks that a later run adds to the log
// rather than replacing it. Each run sees only its own rejections, so a run
// that rewrites the file erases every earlier one — and the cursors have long
// since moved past those ideas, so nothing can rebuild them.
func TestEntriesAccumulateAcrossRuns(t *testing.T) {
	root := t.TempDir()

	runs := []Entry{
		{Idea: "first idea", Source: "src#1", Reason: Irrelevant, Note: "no use for it"},
		{Idea: "second idea", Source: "src#2", Reason: Irrelevant, Note: "still no use"},
		{Idea: "third idea", Source: "src#3", Reason: Irrelevant, Note: "nor this one"},
	}

	var after []string
	for i, e := range runs {
		if err := Append(root, e); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}

		raw, err := os.ReadFile(filepath.Join(root, LogPath))
		if err != nil {
			t.Fatalf("run %d: read the log: %v", i+1, err)
		}
		after = append(after, string(raw))

		// Every earlier run's log is still a prefix of this one. Nothing above
		// the new entry moved, and nothing was rewritten.
		if i > 0 && !strings.HasPrefix(after[i], after[i-1]) {
			t.Fatalf("run %d rewrote what run %d had already written:\n%s", i+1, i, after[i])
		}
	}

	final := after[len(after)-1]
	if n := strings.Count(final, "# Rejections"); n != 1 {
		t.Errorf("the log carries its header %d times; it is written once and appended to after that", n)
	}

	got, err := Parse(final)
	if err != nil {
		t.Fatalf("parse the log: %v", err)
	}
	if len(got) != len(runs) {
		t.Fatalf("%d runs each rejected one idea, log holds %d", len(runs), len(got))
	}
	for i, want := range runs {
		if got[i].Idea != want.Idea {
			t.Errorf("entry %d: got %q, want %q — oldest first", i, got[i].Idea, want.Idea)
		}
	}
}

// TestLogTravelsWithTheCursor checks that one commit carries both. The cursor
// advancing is what makes a rejection permanent: the source item is never
// offered again. If the run advances the cursor and the log does not reach the
// same commit, the idea is gone and nothing says it ever existed. Two commits
// are not enough either — the second can fail after the first has landed.
func TestLogTravelsWithTheCursor(t *testing.T) {
	raw, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read the propose program: %v", err)
	}
	agents, ok := section(string(raw), "## Agents")
	if !ok {
		t.Fatal("the propose program has no agents section")
	}

	// The cursor moves before anything is committed. A commit that precedes
	// the advance leaves the log describing a source position the repository
	// has not reached.
	advance := strings.Index(agents, "Update `last_version`")
	if advance < 0 {
		t.Fatal("the proposer never advances the source cursor")
	}
	commit := strings.Index(agents, "git commit")
	if commit < 0 {
		t.Fatal("the proposer commits nothing")
	}
	if advance > commit {
		t.Error("the proposer commits before it advances the cursor")
	}

	if n := strings.Count(agents, "git commit"); n != 1 {
		t.Errorf("the run makes %d commits; the log and the cursor must land together in one", n)
	}

	add, ok := lineContaining(agents, "git add")
	if !ok {
		t.Fatal("the proposer stages nothing")
	}
	for _, want := range []string{LogPath, ".minions/sources.yaml", ".minions/programs/"} {
		if !strings.Contains(add, want) {
			t.Errorf("the single commit leaves out %q:\n%s", want, add)
		}
	}
}

// section returns the body of the named heading, from the heading to the next
// heading of the same level or to the end. It ignores headings inside a fenced
// block, because the entry template this program shows its agent is itself a
// markdown heading and must not close the section that contains it.
func section(src, heading string) (string, bool) {
	lines := strings.Split(src, "\n")
	closes := strings.Repeat("#", strings.Count(heading, "#")) + " "

	start, fenced := -1, false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}

		switch {
		case start < 0 && trimmed == heading:
			start = i + 1
		case start >= 0 && strings.HasPrefix(trimmed, closes):
			return strings.Join(lines[start:i], "\n"), true
		}
	}
	if start < 0 {
		return "", false
	}
	return strings.Join(lines[start:], "\n"), true
}

// readOne parses the log under root and returns its single entry.
func readOne(t *testing.T, root string) Entry {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(root, LogPath))
	if err != nil {
		t.Fatalf("read the log: %v", err)
	}
	got, err := Parse(string(raw))
	if err != nil {
		t.Fatalf("parse the log: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("wrote 1 entry, read back %d", len(got))
	}
	return got[0]
}

// lineContaining returns the first line of src that contains want.
func lineContaining(src, want string) (string, bool) {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, want) {
			return line, true
		}
	}
	return "", false
}

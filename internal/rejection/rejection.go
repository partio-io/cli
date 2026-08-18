// Package rejection records what the proposer threw away, and why. The run that
// drops an idea also advances the cursor that tracks how far it has read each
// source, so the source item is never offered again. Without a record the idea
// is gone with no trace, and a bar set too high looks exactly like a source that
// has gone quiet. This log is the only surface that tells those two apart.
package rejection

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/partio-io/cli/internal/premise"
)

const (
	// LogPath is where rejections accumulate, relative to the repository root.
	LogPath = ".minions/rejections.md"

	// Marker names an entry and its format version, so a reader finds the
	// entries without guessing at the prose around them.
	Marker = "<!-- partio:rejection:v1 -->"

	// header opens a log file that does not exist yet. It is written once, and
	// every later run appends below it.
	header = "# Rejections\n\n" +
		"Ideas the proposer did not file, and why. Each run appends; nothing here\n" +
		"is ever rewritten. The cursor has already moved past every item below, so\n" +
		"this file is the only record that the idea was seen at all.\n"
)

// Reason is why an idea did not become a proposal.
type Reason string

const (
	// PremiseFailed reports an idea this repository contradicted. It was
	// checked, and a claim it depends on did not survive the check.
	PremiseFailed Reason = "premise-failed"

	// Irrelevant reports a source item that was never about this project. It
	// was never an idea, so no claim was ever checked.
	Irrelevant Reason = "irrelevant"
)

// Reasons is the closed set an entry may carry.
var Reasons = []Reason{PremiseFailed, Irrelevant}

var (
	// ErrNoIdea reports an entry that does not say what was rejected.
	ErrNoIdea = errors.New("rejection: entry names no idea")

	// ErrNoSource reports an entry that does not say where the idea came from.
	ErrNoSource = errors.New("rejection: entry names no source")

	// ErrNoReason reports an entry with no reason, or one outside Reasons.
	ErrNoReason = errors.New("rejection: entry gives no known reason")

	// ErrNoClaim reports a premise-failed entry that does not name the claim
	// that failed, or names one with no evidence behind it.
	ErrNoClaim = errors.New("rejection: entry names no failed claim")

	// ErrNoVerdict reports a premise-failed entry whose verdict is missing,
	// outside premise.Verdicts, or Holds — a claim the tree confirmed did not
	// cause a rejection.
	ErrNoVerdict = errors.New("rejection: entry gives no verdict that rejects")

	// ErrNoFinding reports a premise-failed entry that does not say what the
	// gathered evidence actually showed. Without it the entry asserts the
	// failure instead of recording it.
	ErrNoFinding = errors.New("rejection: entry does not say what the evidence showed")

	// ErrNoNote reports an irrelevant entry that does not say why the item was
	// not about this project.
	ErrNoNote = errors.New("rejection: entry does not say why the item was irrelevant")

	// ErrNotChecked reports an irrelevant entry carrying a claim, a verdict or
	// a finding. An item that was never about this project was never checked,
	// so it has none of those, and an entry that shows them reads as a dropped
	// idea — the one distinction this log exists to keep.
	ErrNotChecked = errors.New("rejection: irrelevant entry carries a check that never ran")
)

// Entry is one rejected idea. A premise-failed entry carries the claim, the
// evidence it named, the verdict and what was found; an irrelevant entry
// carries none of those, because nothing was ever checked. That difference is
// structural, so the two kinds cannot be confused by a reader or by a parser.
type Entry struct {
	// Idea is what was rejected, in one line.
	Idea string

	// Source is where the idea came from — the monitored source and the item
	// within it.
	Source string

	// Reason is why it was rejected.
	Reason Reason

	// Claim is the claim that failed. PremiseFailed only.
	Claim premise.Claim

	// Verdict is what the check returned for that claim. PremiseFailed only.
	Verdict premise.Verdict

	// Found is what the gathered evidence actually showed — the excerpt that
	// contradicted the claim. PremiseFailed only.
	Found string

	// Note is why the item was not about this project. Irrelevant only.
	Note string
}

// claimLine matches the claim of a premise-failed entry. The evidence closes the
// line, matching the grammar of the premise block the claim came from.
var claimLine = regexp.MustCompile("^-\\s+claim:\\s+(.*\\S)\\s+\\[evidence:\\s*`([^`]*)`\\]$")

// fieldLine matches a plain `- name: value` field.
var fieldLine = regexp.MustCompile(`^-\s+([a-z]+):\s+(.*\S)\s*$`)

// Render writes one entry as it appears in the log.
func Render(e Entry) (string, error) {
	idea := oneLine(e.Idea)
	source := oneLine(e.Source)

	if idea == "" {
		return "", ErrNoIdea
	}
	if source == "" {
		return "", fmt.Errorf("%w: %s", ErrNoSource, idea)
	}

	var out strings.Builder
	fmt.Fprintf(&out, "## %s\n\n%s\n\n", idea, Marker)
	fmt.Fprintf(&out, "- source: `%s`\n", source)
	fmt.Fprintf(&out, "- reason: `%s`\n", e.Reason)

	switch e.Reason {
	case PremiseFailed:
		statement := oneLine(e.Claim.Statement)
		evidence := oneLine(e.Claim.Evidence)
		found := oneLine(e.Found)

		if statement == "" || evidence == "" {
			return "", fmt.Errorf("%w: %s", ErrNoClaim, idea)
		}
		if !rejects(e.Verdict) {
			return "", fmt.Errorf("%w: %q on %s", ErrNoVerdict, e.Verdict, idea)
		}
		if found == "" {
			return "", fmt.Errorf("%w: %s", ErrNoFinding, idea)
		}

		fmt.Fprintf(&out, "- claim: %s [evidence: `%s`]\n", statement, evidence)
		fmt.Fprintf(&out, "- verdict: `%s`\n", e.Verdict)
		fmt.Fprintf(&out, "- found: %s\n", found)
	case Irrelevant:
		note := oneLine(e.Note)

		if e.Claim.Statement != "" || e.Claim.Evidence != "" || e.Verdict != "" || e.Found != "" {
			return "", fmt.Errorf("%w: %s", ErrNotChecked, idea)
		}
		if note == "" {
			return "", fmt.Errorf("%w: %s", ErrNoNote, idea)
		}

		fmt.Fprintf(&out, "- note: %s\n", note)
	default:
		return "", fmt.Errorf("%w: %q", ErrNoReason, e.Reason)
	}

	return out.String(), nil
}

// Append adds one entry to the log under root, creating the log if this is the
// first rejection ever recorded. It never rewrites what is already there.
func Append(root string, e Entry) (err error) {
	entry, err := Render(e)
	if err != nil {
		return err
	}

	path := filepath.Join(root, LogPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("rejection: make log directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("rejection: open log: %w", err)
	}
	// A close that fails on a write handle can lose the entry that was just
	// written, which is the silent loss this log exists to stop. Report it.
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("rejection: close log: %w", cerr)
		}
	}()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("rejection: stat log: %w", err)
	}

	out := "\n" + entry
	if info.Size() == 0 {
		out = header + out
	}
	if _, err := f.WriteString(out); err != nil {
		return fmt.Errorf("rejection: append to log: %w", err)
	}
	return nil
}

// Parse reads every entry out of a log, oldest first.
func Parse(log string) ([]Entry, error) {
	var (
		out     []Entry
		current *Entry
		marked  bool
	)

	flush := func() error {
		if current == nil {
			return nil
		}
		if !marked {
			return fmt.Errorf("rejection: entry %q does not carry %s", current.Idea, Marker)
		}
		if current.Source == "" {
			return fmt.Errorf("%w: %s", ErrNoSource, current.Idea)
		}
		if !known(current.Reason) {
			return fmt.Errorf("%w: %s", ErrNoReason, current.Idea)
		}
		out = append(out, *current)
		current = nil
		return nil
	}

	for _, line := range strings.Split(log, "\n") {
		line = strings.TrimSpace(line)

		if heading, ok := strings.CutPrefix(line, "## "); ok {
			if err := flush(); err != nil {
				return nil, err
			}
			idea := strings.TrimSpace(heading)
			if idea == "" {
				return nil, ErrNoIdea
			}
			current = &Entry{Idea: idea}
			marked = false
			continue
		}
		if current == nil {
			continue
		}
		if line == Marker {
			marked = true
			continue
		}
		if m := claimLine.FindStringSubmatch(line); m != nil {
			current.Claim = premise.Claim{Statement: m[1], Evidence: strings.TrimSpace(m[2])}
			continue
		}
		if m := fieldLine.FindStringSubmatch(line); m != nil {
			value := strings.Trim(m[2], "`")
			switch m[1] {
			case "source":
				current.Source = value
			case "reason":
				current.Reason = Reason(value)
			case "verdict":
				current.Verdict = premise.Verdict(value)
			case "found":
				current.Found = value
			case "note":
				current.Note = value
			}
		}
	}

	if err := flush(); err != nil {
		return nil, err
	}
	return out, nil
}

// rejects reports whether v is a verdict that drops an idea. Holds does not:
// a claim the tree confirmed is not why a proposal was rejected.
func rejects(v premise.Verdict) bool {
	return v == premise.Fails || v == premise.Unresolved
}

// known reports whether r is one of Reasons.
func known(r Reason) bool {
	return slices.Contains(Reasons, r)
}

// oneLine trims a field and collapses any line break, so one entry stays one
// readable block and a stray newline cannot forge a second entry.
func oneLine(s string) string {
	return strings.TrimSpace(strings.NewReplacer("\r", " ", "\n", " ").Replace(s))
}

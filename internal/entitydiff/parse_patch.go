package entitydiff

import (
	"log/slog"
	"strconv"
	"strings"
)

// maxLineLen is the longest hunk line a file may contain and still be
// analysed. Longer lines mark generated code, and the cap keeps the
// scanner away from pathological input.
const maxLineLen = 1000

// side is one rebuilt side of a hunk: its lines, and for each line its
// position in the file's section of the patch, used for ordering.
type side struct {
	lines []string
	pos   []int
}

func (s *side) add(text string, pos int) {
	s.lines = append(s.lines, text)
	s.pos = append(s.pos, pos)
}

// hunk holds the old side (context plus removed lines) and the new side
// (context plus added lines) of one diff hunk.
type hunk struct {
	old side
	new side
}

// filePatch is one file's section of a unified diff. OldPath is empty for
// a new file and newPath is empty for a deleted file.
type filePatch struct {
	oldPath  string
	newPath  string
	hunks    []hunk
	longLine bool
}

// patchParser turns a git unified diff into per-file hunks.
type patchParser struct {
	files []*filePatch
	cur   *filePatch

	// Hunk state: lines still expected on each side, and the position
	// counter for the current file.
	inHunk           bool
	oldLeft, newLeft int
	pos              int
}

// parsePatch splits patch into files and hunks. It never fails: anything
// it cannot read is logged at debug and skipped.
func parsePatch(patch string) []*filePatch {
	p := &patchParser{}
	for _, line := range strings.Split(patch, "\n") {
		p.line(strings.TrimSuffix(line, "\r"))
	}
	return p.files
}

func (p *patchParser) line(line string) {
	if p.inHunk {
		if p.hunkLine(line) {
			return
		}
		// Not a hunk line: the hunk is over, read the line as a header.
		p.inHunk = false
	}

	switch {
	case strings.HasPrefix(line, "diff --git "):
		oldPath, newPath := parseGitHeader(strings.TrimPrefix(line, "diff --git "))
		p.startFile(oldPath, newPath)
	case strings.HasPrefix(line, "@@ "):
		p.startHunk(line)
	case strings.HasPrefix(line, "--- "):
		if p.cur == nil || len(p.cur.hunks) > 0 {
			p.startFile("", "")
		}
		p.cur.oldPath = headerPath(strings.TrimPrefix(line, "--- "), "a/")
	case strings.HasPrefix(line, "+++ "):
		if p.cur != nil && len(p.cur.hunks) == 0 {
			p.cur.newPath = headerPath(strings.TrimPrefix(line, "+++ "), "b/")
		}
	case strings.HasPrefix(line, "rename from "):
		if p.cur != nil {
			p.cur.oldPath = unquotePath(strings.TrimPrefix(line, "rename from "))
		}
	case strings.HasPrefix(line, "rename to "):
		if p.cur != nil {
			p.cur.newPath = unquotePath(strings.TrimPrefix(line, "rename to "))
		}
	case strings.HasPrefix(line, "new file mode"):
		if p.cur != nil {
			p.cur.oldPath = ""
		}
	case strings.HasPrefix(line, "deleted file mode"):
		if p.cur != nil {
			p.cur.newPath = ""
		}
	}
	// Every other line (index, mode, similarity, "Binary files ... differ",
	// binary patch data, "\ No newline at end of file") carries nothing
	// the analysis needs.
}

func (p *patchParser) startFile(oldPath, newPath string) {
	p.cur = &filePatch{oldPath: oldPath, newPath: newPath}
	p.files = append(p.files, p.cur)
	p.pos = 0
}

func (p *patchParser) startHunk(line string) {
	if p.cur == nil {
		slog.Debug("entitydiff: hunk outside a file section", "line", line)
		return
	}
	oldCount, newCount, ok := parseHunkHeader(line)
	if !ok {
		slog.Debug("entitydiff: malformed hunk header", "line", line)
		return
	}
	p.cur.hunks = append(p.cur.hunks, hunk{})
	p.inHunk = oldCount > 0 || newCount > 0
	p.oldLeft, p.newLeft = oldCount, newCount
}

// hunkLine consumes one line of a hunk body. It reports false when the
// line does not belong to the hunk.
func (p *patchParser) hunkLine(line string) bool {
	if strings.HasPrefix(line, `\`) {
		return true // "\ No newline at end of file"
	}

	h := &p.cur.hunks[len(p.cur.hunks)-1]
	op, text := byte(' '), ""
	if line != "" {
		// An empty line is a context line whose leading space was lost,
		// e.g. by trimming the end of the stored patch.
		op, text = line[0], line[1:]
	}

	switch {
	case op == ' ' && p.oldLeft > 0 && p.newLeft > 0:
		h.old.add(text, p.pos)
		h.new.add(text, p.pos)
		p.oldLeft--
		p.newLeft--
	case op == '-' && p.oldLeft > 0:
		h.old.add(text, p.pos)
		p.oldLeft--
	case op == '+' && p.newLeft > 0:
		h.new.add(text, p.pos)
		p.newLeft--
	default:
		slog.Debug("entitydiff: unexpected line in hunk", "file", p.cur.newPath)
		return false
	}

	if len(line) > maxLineLen {
		p.cur.longLine = true
	}
	p.pos++
	if p.oldLeft == 0 && p.newLeft == 0 {
		p.inHunk = false
	}
	return true
}

// parseHunkHeader reads the line counts from "@@ -a,b +c,d @@ ...".
// An omitted count means 1.
func parseHunkHeader(line string) (oldCount, newCount int, ok bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 || fields[3] != "@@" ||
		!strings.HasPrefix(fields[1], "-") || !strings.HasPrefix(fields[2], "+") {
		return 0, 0, false
	}
	oldCount, ok = rangeCount(fields[1][1:])
	if !ok {
		return 0, 0, false
	}
	newCount, ok = rangeCount(fields[2][1:])
	return oldCount, newCount, ok
}

func rangeCount(r string) (int, bool) {
	start, count, found := strings.Cut(r, ",")
	if _, err := strconv.Atoi(start); err != nil {
		return 0, false
	}
	if !found {
		return 1, true
	}
	n, err := strconv.Atoi(count)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// parseGitHeader reads the two paths of "diff --git a/X b/Y". The
// ---/+++ and rename lines that follow override them when present.
func parseGitHeader(rest string) (oldPath, newPath string) {
	if strings.HasPrefix(rest, `"`) {
		// Quoted paths: the ---/+++ or rename lines name the file.
		return "", ""
	}
	// For an unrenamed file both halves are equal, which finds the split
	// even when the path itself contains " b/".
	if n := len(rest); n%2 == 1 {
		mid := n / 2
		if rest[mid] == ' ' && strings.HasPrefix(rest, "a/") && strings.HasPrefix(rest[mid+1:], "b/") &&
			rest[2:mid] == rest[mid+3:] {
			return rest[2:mid], rest[mid+3:]
		}
	}
	a, b, found := strings.Cut(rest, " b/")
	if !found {
		return "", ""
	}
	return strings.TrimPrefix(a, "a/"), b
}

// headerPath reads the path of a ---/+++ line. "/dev/null" gives "".
func headerPath(raw, prefix string) string {
	raw = strings.TrimSuffix(raw, "\t")
	if raw == "/dev/null" {
		return ""
	}
	return strings.TrimPrefix(unquotePath(raw), prefix)
}

// unquotePath undoes git's C-style quoting of unusual paths.
func unquotePath(raw string) string {
	if strings.HasPrefix(raw, `"`) {
		if s, err := strconv.Unquote(raw); err == nil {
			return s
		}
	}
	return raw
}

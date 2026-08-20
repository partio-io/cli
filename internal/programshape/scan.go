package programshape

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Scan is the result of reading a minions programs directory.
type Scan struct {
	// Programs are the program file names read, sorted.
	Programs []string
	// Unreachable maps a program file name to the sections whose instructions
	// never reach the model. A program in good shape is absent from the map.
	Unreachable map[string][]Finding
}

// ScanDir runs Check over every .md program in dir.
func ScanDir(dir string) (Scan, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Scan{}, fmt.Errorf("read programs directory %s: %w", dir, err)
	}

	scan := Scan{Unreachable: make(map[string][]Finding)}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}

		src, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return Scan{}, fmt.Errorf("read program %s: %w", name, err)
		}

		scan.Programs = append(scan.Programs, name)
		if findings := Check(string(src)); len(findings) > 0 {
			scan.Unreachable[name] = findings
		}
	}
	slices.Sort(scan.Programs)

	return scan, nil
}

// Report renders one program's findings as a failure message. It names the
// program file and every dropped heading, so the repair needs no bisecting of
// the program format rules.
func Report(path string, findings []Finding) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s: instructions do not reach the model\n", path)
	for _, f := range findings {
		fmt.Fprintf(&b, "\tthe dropped section %q carries %s\n", f.Heading, f.Reason)
	}
	b.WriteString("\tThe program parser keeps only the H1 prose, ## Context, ## Planner and\n" +
		"\t## Agents. Move these instructions into an ## Agents sub-section, or record\n" +
		"\tthe program in knownUnreachable with a reason.")

	return b.String()
}

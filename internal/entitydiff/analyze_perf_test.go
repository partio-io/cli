package entitydiff

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// perfBudget is the longest Analyze may take on one checkpoint.
const perfBudget = 100 * time.Millisecond

// largePatch builds a synthetic patch of files Go and TS files, half
// each, of about linesPerFile lines apiece. Every hunk mixes context,
// added, removed, re-signed and renamed declarations with body lines, so
// that every part of the analysis runs.
func largePatch(files, linesPerFile int) string {
	var b strings.Builder
	for f := 0; f < files; f++ {
		ext := ".go"
		if f%2 == 1 {
			ext = ".ts"
		}
		path := fmt.Sprintf("pkg%d/file%d%s", f/10, f, ext)
		fmt.Fprintf(&b, "diff --git a/%s b/%s\nindex 1111111..2222222 100644\n--- a/%s\n+++ b/%s\n", path, path, path, path)
		for line, d := 0, 0; line < linesPerFile; d++ {
			fmt.Fprintf(&b, "@@ -%d,12 +%d,12 @@\n", line+1, line+1)
			if ext == ".go" {
				fmt.Fprintf(&b, " func Keep%d(a int) error {\n", d)
				b.WriteString(" \tx := a * 2\n")
				fmt.Fprintf(&b, "-\treturn fmt.Errorf(\"old %%d\", x)\n+\treturn fmt.Errorf(\"new %%d\", x)\n")
				b.WriteString(" }\n")
				fmt.Fprintf(&b, "-func (s *Store) Load%d(id string) error {\n+func (s *Store) Load%d(id string, opts Options) error {\n", d, d)
				fmt.Fprintf(&b, "-func old%d(n int) int {\n+func new%d(n int) int {\n", d, d)
				fmt.Fprintf(&b, "+type Added%d struct {\n", d)
				fmt.Fprintf(&b, "-type Gone%d interface {\n", d)
				b.WriteString(" \treturn n\n")
			} else {
				fmt.Fprintf(&b, " export function keep%d(a: number): number {\n", d)
				b.WriteString(" if (a > 0) {\n")
				fmt.Fprintf(&b, "-  return a;\n+  return a * 2;\n")
				b.WriteString(" }\n")
				fmt.Fprintf(&b, "-export async function load%d(id: string) {\n+export async function load%d(id: string, opts?: Options) {\n", d, d)
				fmt.Fprintf(&b, "-export class Old%d {\n+export class New%d {\n", d, d)
				fmt.Fprintf(&b, "+export interface Added%d {\n", d)
				fmt.Fprintf(&b, "-export const gone%d = (x: number) => {\n", d)
				b.WriteString("   return x;\n")
			}
			line += 14
		}
	}
	return b.String()
}

func BenchmarkAnalyze(b *testing.B) {
	p := largePatch(200, 100)
	b.SetBytes(int64(len(p)))
	b.ResetTimer()
	for b.Loop() {
		Analyze(p)
	}
}

// TestAnalyzePerformanceBudget fails when Analyze takes longer than the
// 100 ms budget on a large checkpoint: about 200 files and 20k lines. The
// best of 3 runs is used, so one slow run on a busy machine does not fail
// it. It is skipped under -short and -race, where timing means nothing.
func TestAnalyzePerformanceBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("timing guard skipped in -short mode")
	}
	if raceEnabled {
		t.Skip("timing guard skipped under the race detector")
	}
	p := largePatch(200, 100)
	if n := strings.Count(p, "\n"); n < 20000 {
		t.Fatalf("fixture has %d lines, want at least 20000", n)
	}
	files := Analyze(p)
	if len(files) != 200 {
		t.Fatalf("Analyze() reported %d files, want 200", len(files))
	}

	best := time.Duration(1<<63 - 1)
	for range 3 {
		start := time.Now()
		Analyze(p)
		best = min(best, time.Since(start))
	}
	if best > perfBudget {
		t.Errorf("Analyze() best of 3 took %v, budget is %v", best, perfBudget)
	}
	t.Logf("Analyze() best of 3: %v", best)
}

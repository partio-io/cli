package entitydiff

import (
	"strings"
	"testing"
)

// patch joins lines into a patch with no trailing newline, the way the
// stored diff.patch reads back.
func patch(lines ...string) string {
	return strings.Join(lines, "\n")
}

// render prints Analyze output compactly: one line per file, then one
// indented line per change.
func render(files []FileChanges) string {
	var b strings.Builder
	for _, f := range files {
		b.WriteString(f.DisplayPath() + "\n")
		for _, c := range f.Changes {
			switch c.Kind {
			case Added:
				b.WriteString("  added " + c.New.Name + ": " + c.New.Signature + "\n")
			case Removed:
				b.WriteString("  removed " + c.Old.Name + ": " + c.Old.Signature + "\n")
			case Renamed:
				b.WriteString("  renamed " + c.Old.Name + " → " + c.New.Name + ": " + c.New.Signature + "\n")
			case SignatureChanged:
				b.WriteString("  signature changed " + c.New.Name + ": " + c.Old.Signature + " => " + c.New.Signature + "\n")
			}
		}
	}
	return b.String()
}

func fileHeader(path string) []string {
	return []string{
		"diff --git a/" + path + " b/" + path,
		"index 1111111..2222222 100644",
		"--- a/" + path,
		"+++ b/" + path,
	}
}

func with(header []string, lines ...string) []string {
	return append(append([]string{}, header...), lines...)
}

func TestAnalyzeGo(t *testing.T) {
	tests := []struct {
		name  string
		patch string
		want  string
	}{
		{
			name: "added function",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,3 +1,7 @@",
				" package a",
				" ",
				" func Old() {}",
				"+",
				"+func New(x int) error {",
				"+\treturn nil",
				"+}",
			)...),
			want: "a.go\n  added New: func New(x int) error\n",
		},
		{
			name: "removed function and type",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,6 +1,2 @@",
				" package a",
				" ",
				"-type Legacy struct {",
				"-\tX int",
				"-}",
				"-func Gone() {}",
			)...),
			want: "a.go\n  removed Legacy: type Legacy struct\n  removed Gone: func Gone()\n",
		},
		{
			name: "signature changed on the declaration line",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,3 +1,3 @@",
				" package a",
				"-func Read(id string) error {",
				"+func Read(id string) (*Data, error) {",
				" \treturn nil",
			)...),
			want: "a.go\n  signature changed Read: func Read(id string) error => func Read(id string) (*Data, error)\n",
		},
		{
			name: "methods with pointer and value receivers",
			patch: patch(with(fileHeader("store.go"),
				"@@ -1,4 +1,7 @@",
				" package a",
				"-func (s *Store) Write(a int) error {",
				"+func (s *Store) Write(a, b int) error {",
				" \treturn nil",
				" }",
				"+",
				"+func (t T) String() string { return \"\" }",
				"+func (l *List[K, V]) Len() int { return 0 }",
			)...),
			want: "store.go\n" +
				"  signature changed (*Store).Write: func (s *Store) Write(a int) error => func (s *Store) Write(a, b int) error\n" +
				"  added T.String: func (t T) String() string\n" +
				"  added (*List).Len: func (l *List[K, V]) Len() int\n",
		},
		{
			name: "grouped type block, const and var ignored",
			patch: patch(with(fileHeader("types.go"),
				"@@ -1,5 +1,9 @@",
				" package a",
				" type (",
				"-\tA struct {",
				"+\tA interface {",
				" \t}",
				"+\tB = A",
				"+\tC[T any] func(T) error",
				" )",
				"+const X = 1",
				"+var Y = 2",
			)...),
			want: "types.go\n" +
				"  signature changed A: type A struct => type A interface\n" +
				"  added B: type B = A\n" +
				"  added C: type C[T any] func(T) error\n",
		},
		{
			name: "change on a continuation line with the func line as context",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,4 +1,5 @@",
				" func F(",
				" \ta int,",
				"+\tb int,",
				" ) error {",
				" \treturn nil",
			)...),
			want: "a.go\n  signature changed F: func F(a int) error => func F(a int, b int) error\n",
		},
		{
			name: "result type literal does not end the signature",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,1 +1,1 @@",
				"-func F() interface{ M() } {",
				"+func F() interface{ N() } {",
			)...),
			want: "a.go\n  signature changed F: func F() interface{ M() } => func F() interface{ N() }\n",
		},
		{
			name: "whitespace-only reflow is not reported",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,2 +1,5 @@",
				"-func F(a int, b int) error { // doc",
				"+func F(",
				"+\ta int,",
				"+\tb int,",
				"+) error {",
				" \treturn nil",
			)...),
			want: "",
		},
		{
			name: "body-only edit is not reported",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,3 +1,4 @@",
				" func F() error {",
				"-\treturn nil",
				"+\tf := func(x int) {}",
				"+\treturn f",
				" }",
			)...),
			want: "",
		},
		{
			name: "declaration cut off at a hunk boundary is dropped",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,3 +1,2 @@",
				" package a",
				"-func G(a int) {}",
				"-func F(a int) {",
				"+func F(a int,",
			)...),
			want: "a.go\n  removed G: func G(a int)\n",
		},
		{
			name: "test files are analysed",
			patch: patch(with(fileHeader("a_test.go"),
				"@@ -1,1 +1,2 @@",
				" package a",
				"+func helper(t *testing.T) {}",
			)...),
			want: "a_test.go\n  added helper: func helper(t *testing.T)\n",
		},
		{
			name: "files and entities keep diff order",
			patch: patch(append(with(fileHeader("b.go"),
				"@@ -1,2 +1,2 @@",
				"-func Z() {}",
				"+func A() {}",
				" func M() {}",
				"@@ -10,1 +10,2 @@",
				" func N() {}",
				"+func B() {}",
			), with(fileHeader("a.go"),
				"@@ -1,0 +1,1 @@",
				"+type T int",
			)...)...),
			want: "b.go\n  removed Z: func Z()\n  added A: func A()\n  added B: func B()\n" +
				"a.go\n  added T: type T int\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := render(Analyze(tt.patch)); got != tt.want {
				t.Errorf("Analyze() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestAnalyzePatchShapes(t *testing.T) {
	longLine := "+// " + strings.Repeat("x", maxLineLen)

	tests := []struct {
		name  string
		patch string
		want  string
	}{
		{
			name: "new file",
			patch: patch(
				"diff --git a/new.go b/new.go",
				"new file mode 100644",
				"index 0000000..1111111",
				"--- /dev/null",
				"+++ b/new.go",
				"@@ -0,0 +1,4 @@",
				"+package a",
				"+",
				"+func F() {}",
				"+type T struct{}",
			),
			want: "new.go\n  added F: func F()\n  added T: type T struct\n",
		},
		{
			name: "deleted file",
			patch: patch(
				"diff --git a/old.go b/old.go",
				"deleted file mode 100644",
				"index 1111111..0000000",
				"--- a/old.go",
				"+++ /dev/null",
				"@@ -1,2 +0,0 @@",
				"-package a",
				"-func F() {}",
			),
			want: "old.go\n  removed F: func F()\n",
		},
		{
			name: "rename with rename headers",
			patch: patch(
				"diff --git a/old.go b/new.go",
				"similarity index 90%",
				"rename from old.go",
				"rename to new.go",
				"index 1111111..2222222 100644",
				"--- a/old.go",
				"+++ b/new.go",
				"@@ -1,3 +1,3 @@",
				" package a",
				" ",
				"-func F() {}",
				"+func F(x int) {}",
			),
			want: "old.go → new.go\n  signature changed F: func F() => func F(x int)\n",
		},
		{
			name: "pure rename shows nothing",
			patch: patch(
				"diff --git a/old.go b/new.go",
				"similarity index 100%",
				"rename from old.go",
				"rename to new.go",
			),
			want: "",
		},
		{
			name: "rename as delete plus add",
			patch: patch(
				"diff --git a/old.go b/old.go",
				"deleted file mode 100644",
				"index 1111111..0000000",
				"--- a/old.go",
				"+++ /dev/null",
				"@@ -1,7 +0,0 @@",
				"-package a",
				"-func A() int {",
				"-\treturn 1",
				"-}",
				"-func B() int {",
				"-\treturn 2",
				"-}",
				"diff --git a/new.go b/new.go",
				"new file mode 100644",
				"index 0000000..1111111",
				"--- /dev/null",
				"+++ b/new.go",
				"@@ -0,0 +1,10 @@",
				"+package a",
				"+func A() int {",
				"+\treturn 1",
				"+}",
				"+func B() int {",
				"+\treturn 2",
				"+}",
				"+func C() int {",
				"+\treturn 3",
				"+}",
			),
			want: "old.go → new.go\n  added C: func C() int\n",
		},
		{
			name: "unrelated delete and add stay separate",
			patch: patch(
				"diff --git a/old.go b/old.go",
				"deleted file mode 100644",
				"--- a/old.go",
				"+++ /dev/null",
				"@@ -1,1 +0,0 @@",
				"-func Old() {}",
				"diff --git a/new.go b/new.go",
				"new file mode 100644",
				"--- /dev/null",
				"+++ b/new.go",
				"@@ -0,0 +1,1 @@",
				"+func New(x string) {}",
			),
			want: "old.go\n  removed Old: func Old()\nnew.go\n  added New: func New(x string)\n",
		},
		{
			name: "binary block, no-newline markers and no trailing newline",
			patch: patch(
				"diff --git a/img.png b/img.png",
				"index 1111111..2222222 100644",
				"Binary files a/img.png and b/img.png differ",
				"diff --git a/a.go b/a.go",
				"index 1111111..2222222 100644",
				"--- a/a.go",
				"+++ b/a.go",
				"@@ -1,2 +1,2 @@",
				" package a",
				"-func F() {}",
				`\ No newline at end of file`,
				"+func F(x int) {}",
				`\ No newline at end of file`,
			),
			want: "a.go\n  signature changed F: func F() => func F(x int)\n",
		},
		{
			name: "trailing context lost to trimming",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,3 +1,4 @@",
				"+func F() {}",
				" func G() {}",
			)...),
			want: "a.go\n  added F: func F()\n",
		},
		{
			name: "removed line that looks like a header",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,2 +1,1 @@",
				"--- a/x",
				" func G() {}",
				"@@ -9,0 +9,1 @@",
				"+func H() {}",
			)...),
			want: "a.go\n  added H: func H()\n",
		},
		{
			name: "unsupported extensions",
			patch: patch(append(with(fileHeader("README.md"),
				"@@ -1,0 +1,1 @@",
				"+func F() {}",
			), with(fileHeader("main.py"),
				"@@ -1,0 +1,1 @@",
				"+func F() {}",
			)...)...),
			want: "",
		},
		{
			name: "file with an over-long line is skipped",
			patch: patch(with(fileHeader("gen.go"),
				"@@ -1,0 +1,2 @@",
				"+func F() {}",
				longLine,
			)...),
			want: "",
		},
		{
			name:  "empty patch",
			patch: "",
			want:  "",
		},
		{
			name:  "malformed patch",
			patch: patch("not a diff", "@@ nonsense @@", "+func F() {}", "--- ", "@@ -x +y @@", "+func G() {}"),
			want:  "",
		},
		{
			name: "malformed hunk does not hide the next file",
			patch: patch(append(with(fileHeader("a.go"),
				"@@ -1,9 +1,9 @@",
				"+func F(",
			), with(fileHeader("b.go"),
				"@@ -1,0 +1,1 @@",
				"+func G() {}",
			)...)...),
			want: "b.go\n  added G: func G()\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := render(Analyze(tt.patch)); got != tt.want {
				t.Errorf("Analyze() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

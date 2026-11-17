package entitydiff

import "testing"

func TestAnalyzeRenames(t *testing.T) {
	tests := []struct {
		name  string
		patch string
		want  string
	}{
		{
			name: "go function renamed",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,4 +1,4 @@",
				" package a",
				" ",
				"-func oldName(id string) (int, error) {",
				"+func newName(id string) (int, error) {",
				" \treturn 0, nil",
			)...),
			want: "a.go\n  renamed oldName → newName: func newName(id string) (int, error)\n",
		},
		{
			name: "go method renamed keeps receiver",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,2 +1,2 @@",
				"-func (s *Store) Load(id string) error {",
				"+func (s *Store) Fetch(id string) error {",
			)...),
			want: "a.go\n  renamed (*Store).Load → (*Store).Fetch: func (s *Store) Fetch(id string) error\n",
		},
		{
			name: "go type renamed in a group",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,4 +1,4 @@",
				" type (",
				"-\tLegacy struct {",
				"+\tModern struct {",
				" \t\tX int",
			)...),
			want: "a.go\n  renamed Legacy → Modern: type Modern struct\n",
		},
		{
			name: "ts class and arrow binding renamed",
			patch: patch(with(fileHeader("web/a.ts"),
				"@@ -1,4 +1,4 @@",
				"-export class Loader {",
				"+export class Fetcher {",
				" }",
				"-export const load = async (id: string): Promise<void> => {",
				"+export const fetch = async (id: string): Promise<void> => {",
			)...),
			want: "web/a.ts\n  renamed Loader → Fetcher: export class Fetcher\n  renamed load → fetch: export const fetch = async (id: string): Promise<void>\n",
		},
		{
			name: "rename survives whitespace reflow",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,3 +1,3 @@",
				"-func Old(a int,",
				"-\tb int) {",
				"+func New(a int, b int) {",
			)...),
			want: "a.go\n  renamed Old → New: func New(a int, b int)\n",
		},
		{
			name: "rename with changed signature stays removed plus added",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,1 +1,1 @@",
				"-func Old(a int) {}",
				"+func New(a string) {}",
			)...),
			want: "a.go\n  removed Old: func Old(a int)\n  added New: func New(a string)\n",
		},
		{
			name: "kind mismatch stays removed plus added",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,1 +1,1 @@",
				"-class Thing {",
				"+interface Other {",
			)...),
			want: "a.ts\n  removed Thing: class Thing\n  added Other: interface Other\n",
		},
		{
			name: "modifier change stays removed plus added",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,1 +1,1 @@",
				"-function load(id: string) {",
				"+export function fetch(id: string) {",
			)...),
			want: "a.ts\n  removed load: function load(id: string)\n  added fetch: export function fetch(id: string)\n",
		},
		{
			name: "method moved to another receiver stays removed plus added",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,1 +1,1 @@",
				"-func (s *A) Run() {}",
				"+func (s *B) Go() {}",
			)...),
			want: "a.go\n  removed (*A).Run: func (s *A) Run()\n  added (*B).Go: func (s *B) Go()\n",
		},
		{
			name: "two candidates stay removed plus added",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,1 +1,2 @@",
				"-func Old() {}",
				"+func NewA() {}",
				"+func NewB() {}",
			)...),
			want: "a.go\n  removed Old: func Old()\n  added NewA: func NewA()\n  added NewB: func NewB()\n",
		},
		{
			name: "two removed for one added stay removed plus added",
			patch: patch(with(fileHeader("a.go"),
				"@@ -1,2 +1,1 @@",
				"-func OldA() {}",
				"-func OldB() {}",
				"+func New() {}",
			)...),
			want: "a.go\n  removed OldA: func OldA()\n  removed OldB: func OldB()\n  added New: func New()\n",
		},
		{
			name: "renames across files are not paired",
			patch: patch(append(with(fileHeader("a.go"),
				"@@ -1,1 +1,0 @@",
				"-func Old() {}",
			), with(fileHeader("b.go"),
				"@@ -1,0 +1,1 @@",
				"+func New() {}",
			)...)...),
			want: "a.go\n  removed Old: func Old()\nb.go\n  added New: func New()\n",
		},
		{
			name: "renamed entity in a renamed file",
			patch: patch(
				"diff --git a/old.go b/new.go",
				"similarity index 90%",
				"rename from old.go",
				"rename to new.go",
				"--- a/old.go",
				"+++ b/new.go",
				"@@ -1,3 +1,3 @@",
				" package a",
				" ",
				"-func F(x int) {}",
				"+func G(x int) {}",
			),
			want: "old.go → new.go\n  renamed F → G: func G(x int)\n",
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

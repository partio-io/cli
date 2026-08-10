package checksverdict

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// The boundary a repair session is held to: a test the pull request
// added is the minion's own work and may be corrected; every other
// test predates the pull request and is out of bounds. Deciding it
// here, from the diff, is what stops the session from reaching green
// by deleting an assertion someone else wrote.
func TestAddedTestFiles(t *testing.T) {
	tests := []struct {
		name       string
		nameStatus string
		want       []string
	}{
		{
			name:       "no diff at all",
			nameStatus: "",
			want:       nil,
		},
		{
			name:       "a test the pull request added",
			nameStatus: "A\tinternal/foo/bar_test.go\n",
			want:       []string{"internal/foo/bar_test.go"},
		},
		{
			name: "implementation and its new test",
			nameStatus: "M\tinternal/foo/bar.go\n" +
				"A\tinternal/foo/bar_test.go\n",
			want: []string{"internal/foo/bar_test.go"},
		},
		// The common case the rule protects. A test that existed
		// before the pull request stays out of bounds even when the
		// pull request edited it — the assertions in it are not the
		// minion's to remove.
		{
			name:       "a test the pull request only modified",
			nameStatus: "M\tinternal/foo/existing_test.go\n",
			want:       nil,
		},
		{
			name:       "a test the pull request deleted",
			nameStatus: "D\tinternal/foo/gone_test.go\n",
			want:       nil,
		},
		// A renamed file is old content at a new path. Its assertions
		// predate the pull request, so it is not an addition.
		{
			name:       "a test the pull request renamed",
			nameStatus: "R100\tinternal/old/bar_test.go\tinternal/new/bar_test.go\n",
			want:       nil,
		},
		{
			name:       "a test the pull request copied",
			nameStatus: "C085\tinternal/old/bar_test.go\tinternal/new/bar_test.go\n",
			want:       nil,
		},
		{
			name:       "an added file that is not a test",
			nameStatus: "A\tinternal/foo/baz.go\n",
			want:       nil,
		},
		// "_test.go" is the whole rule: a file merely named after
		// testing is implementation and needs no exemption.
		{
			name: "files that only look like tests",
			nameStatus: "A\tinternal/foo/test.go\n" +
				"A\tinternal/foo/testing.go\n" +
				"A\tinternal/foo/my_test_helper.go\n",
			want: nil,
		},
		{
			name: "several additions, sorted",
			nameStatus: "A\tinternal/z/z_test.go\n" +
				"A\tinternal/a/a_test.go\n" +
				"A\tinternal/m/m_test.go\n",
			want: []string{"internal/a/a_test.go", "internal/m/m_test.go", "internal/z/z_test.go"},
		},
		{
			name: "the same file added twice across commits",
			nameStatus: "A\tinternal/foo/bar_test.go\n" +
				"A\tinternal/foo/bar_test.go\n",
			want: []string{"internal/foo/bar_test.go"},
		},
		{
			name: "blank and malformed lines are ignored",
			nameStatus: "\n" +
				"A\n" +
				"   \n" +
				"A\tinternal/foo/bar_test.go\n",
			want: []string{"internal/foo/bar_test.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddedTestFiles(tt.nameStatus)

			if !slices.Equal(got, tt.want) {
				t.Errorf("AddedTestFiles = %q, want %q", got, tt.want)
			}
		})
	}
}

// The session reads this file to learn which tests it may touch, so
// one path per line and nothing else.
func TestWriteAddedTests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "added-tests.txt")

	if err := WriteAddedTests(path, []string{"internal/a/a_test.go", "internal/b/b_test.go"}); err != nil {
		t.Fatalf("WriteAddedTests: %v", err)
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read the list back: %v", err)
	}
	if got, want := string(data), "internal/a/a_test.go\ninternal/b/b_test.go\n"; got != want {
		t.Errorf("wrote %q, want %q", got, want)
	}
}

// A pull request that added no tests still gets a file. "No test is
// yours to edit" and "the list was never written" have to look
// different to the session, or a crash upstream would read as a licence
// to edit every test in the repository.
func TestWriteAddedTestsWritesAnEmptyList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "added-tests.txt")

	if err := WriteAddedTests(path, nil); err != nil {
		t.Fatalf("WriteAddedTests: %v", err)
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("the empty list was not written: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("wrote %q for an empty list, want nothing", data)
	}
}

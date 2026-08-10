// Command minion-apply-fix turns a repair session's patch into one
// pushed commit. It fetches the pull request's real head, detaches a
// worktree onto it, applies the patch, proves the result with the
// repository's own lint and test targets, and pushes a commit marked
// with the check that produced it.
//
// It writes pushed to $GITHUB_OUTPUT. A patch that does not apply, or
// that leaves the checks red, is a spent round and not an error: the
// command says so and exits 0, and the first verdict stands.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/partio-io/cli/internal/patchapply"
)

// maxBodyBytes bounds the commit body taken from the session's fix
// summary. A commit body is read by people.
const maxBodyBytes = 300

func main() {
	var (
		pr       = flag.Int("pr", 0, "pull request number (required)")
		check    = flag.String("check", "", "check that produced the patch, e.g. dead-code (required)")
		patch    = flag.String("patch", ".minion-audit/fix.patch", "patch the repair session exported")
		summary  = flag.String("summary", ".minion-audit/fix-summary.txt", "summary used as the commit body")
		worktree = flag.String("worktree", "", "where to detach the pull request head")
	)
	flag.Parse()
	if *pr == 0 || *check == "" {
		fmt.Fprintln(os.Stderr,
			"usage: minion-apply-fix --pr <number> --check <name> [--patch <path>] [--summary <path>] [--worktree <dir>]")
		os.Exit(2)
	}
	repo := os.Getenv("GITHUB_REPOSITORY")
	token := os.Getenv("GH_TOKEN")
	if repo == "" || token == "" {
		fmt.Fprintln(os.Stderr, "minion-apply-fix: GITHUB_REPOSITORY and GH_TOKEN must be set")
		os.Exit(2)
	}
	api := os.Getenv("GITHUB_API_URL")
	if api == "" {
		api = "https://api.github.com"
	}

	res, runErr := patchapply.Run(patchapply.Config{
		PatchPath:   *patch,
		WorktreeDir: *worktree,
		Body:        body(*summary),
		Repo:        repo,
		PRNumber:    *pr,
		Check:       *check,
		Token:       token,
		APIBaseURL:  api,
	})
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "minion-apply-fix: %v\n", runErr)
	}
	if res.Pushed {
		fmt.Printf("minion-apply-fix: %s\n", res.Reason)
	} else if res.Reason != "" {
		fmt.Printf("minion-apply-fix: pushed nothing. %s\n", res.Reason)
	}

	// The output goes out before the exit code, because the job reads
	// pushed on every path, including the failed one.
	if err := writeOutputs(res); err != nil {
		fmt.Fprintf(os.Stderr, "minion-apply-fix: %v\n", err)
		os.Exit(1)
	}
	if runErr != nil {
		os.Exit(1)
	}
}

// body returns the commit body, taken from the session's fix summary.
// A missing summary is normal, and gives an empty body.
func body(path string) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(data))
	if len(text) > maxBodyBytes {
		text = strings.ToValidUTF8(text[:maxBodyBytes], "")
	}
	return text
}

// writeOutputs appends the outcome to the step's $GITHUB_OUTPUT file.
func writeOutputs(res patchapply.Result) error {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open GITHUB_OUTPUT: %w", err)
	}
	_, writeErr := fmt.Fprintf(f, "pushed=%t\n", res.Pushed)
	closeErr := f.Close()
	if writeErr != nil {
		return fmt.Errorf("write GITHUB_OUTPUT: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close GITHUB_OUTPUT: %w", closeErr)
	}
	return nil
}

package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/git"
	"github.com/partio-io/cli/internal/github"
)

const defaultGitHubAPIURL = "https://api.github.com"

func newSyncForkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync-fork <pr>",
		Short: "Copy a fork pull request's checkpoints into the local checkpoint branch",
		Long: `Reads the Partio-Checkpoint trailers of a fork pull request's commits, fetches
the fork's ` + git.CheckpointBranch + ` branch as data, and copies only the referenced
checkpoints into the local ` + git.CheckpointBranch + ` branch in one commit, so that
"partio explain <id>" resolves them.

<pr> is a pull request URL (https://github.com/<owner>/<repo>/pull/<n>) or
<owner>/<repo>#<n>. The fork is never written to and nothing is pushed.

A pull request with nothing to sync (not from a fork, a deleted fork, a fork
without a checkpoint branch, or no checkpoint trailers) exits successfully
with a one-line reason.

The GitHub token is read from GH_TOKEN, then GITHUB_TOKEN; without one, requests
are unauthenticated. The API root is read from GITHUB_API_URL (default
` + defaultGitHubAPIURL + `).`,
		Args: cobra.ExactArgs(1),
		RunE: runSyncFork,
	}
}

func runSyncFork(_ *cobra.Command, args []string) error {
	ref, err := parsePRRef(args[0])
	if err != nil {
		return err
	}

	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	client := newGitHubClientFromEnv()
	pr, err := client.GetPullHead(ref.FullName(), ref.Number)
	if err != nil {
		return err
	}
	// A deleted fork shows up as a null or missing head repository, or one
	// without a clone URL; all are treated alike.
	if pr.Head.Repo == nil || pr.Head.Repo.CloneURL == "" {
		fmt.Printf("nothing to sync: %s has no head repository (the fork was deleted)\n", ref)
		return nil
	}
	head := *pr.Head.Repo
	base := pr.Base.Repo.FullName
	slog.Debug("resolved pull request", "pr", ref.String(), "head_repo", head.FullName, "head_sha", pr.Head.SHA, "base_repo", base, "commits", pr.Commits)
	if strings.EqualFold(head.FullName, base) {
		fmt.Printf("nothing to sync: %s is not from a fork (head and base are both %s)\n", ref, base)
		return nil
	}

	report := syncForkReport{BaseRepo: base, TotalCommits: pr.Commits}
	if origin, ok := originRepo(repoRoot); !ok {
		slog.Debug("origin names no GitHub repository; not comparing it with the pull request's base", "base_repo", base)
	} else if !strings.EqualFold(origin, base) {
		report.OriginRepo = origin
		slog.Warn("the pull request's base repository differs from origin; checkpoints are still synced into the local branch", "base_repo", base, "origin", origin)
	}
	if report.truncated() {
		slog.Warn(fmt.Sprintf("pull request has more than %d commits; commits after the first %d were not scanned", github.MaxPullCommits, github.MaxPullCommits), "commits", pr.Commits)
	}

	commits, err := client.ListPullCommits(ref.FullName(), ref.Number)
	if err != nil {
		return err
	}
	messages := make([]string, 0, len(commits))
	for _, c := range commits {
		messages = append(messages, c.Commit.Message)
	}
	ids, err := checkpointIDsFromMessages(repoRoot, messages)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		fmt.Printf("nothing to sync: no commit of %s carries a valid %s trailer\n", ref, checkpointTrailerKey)
		return nil
	}

	hasBranch, err := forkHasCheckpointBranch(repoRoot, head.CloneURL)
	if err != nil {
		return err
	}
	if !hasBranch {
		fmt.Printf("nothing to sync: fork %s has no %s branch\n", head.FullName, git.CheckpointBranch)
		return nil
	}

	forkRef, cleanup, err := fetchForkBranch(repoRoot, head.CloneURL)
	defer cleanup()
	if err != nil {
		return err
	}

	results, err := checkpoint.NewStore(repoRoot).Import(forkRef, ids, func(synced []string) string {
		return syncForkCommitMessage(ref, head.FullName, synced)
	})
	if err != nil {
		return fmt.Errorf("importing checkpoints: %w", err)
	}

	printSyncForkSummary(results, report)
	return nil
}

// syncForkReport is what the sync summary says about the pull request
// beyond the per-checkpoint outcomes.
type syncForkReport struct {
	BaseRepo     string // the pull request's base repository
	OriginRepo   string // origin's owner/repo when it differs from BaseRepo
	TotalCommits int    // the pull request's commit count
}

// truncated reports whether the pull request has commits that the commit
// listing did not return.
func (r syncForkReport) truncated() bool { return r.TotalCommits > github.MaxPullCommits }

// newGitHubClientFromEnv builds a GitHub client from GITHUB_API_URL and
// GH_TOKEN (falling back to GITHUB_TOKEN).
func newGitHubClientFromEnv() github.Client {
	baseURL := strings.TrimRight(os.Getenv("GITHUB_API_URL"), "/")
	if baseURL == "" {
		baseURL = defaultGitHubAPIURL
	}
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		slog.Debug("no GH_TOKEN or GITHUB_TOKEN set; GitHub API requests are unauthenticated (60 per hour)")
	}
	return github.Client{BaseURL: baseURL, Token: token}
}

// syncForkCommitMessage is the message of the commit that records the
// checkpoints synced from a fork pull request.
func syncForkCommitMessage(ref prRef, headRepo string, synced []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "sync-fork: %s (%d checkpoints)\n\n", ref, len(synced))
	fmt.Fprintf(&b, "Head repository: %s\n\nCheckpoints:\n", headRepo)
	for _, id := range synced {
		fmt.Fprintf(&b, "- %s\n", id)
	}
	return b.String()
}

// printSyncForkSummary warns about every checkpoint that was not synced and
// prints the per-outcome counts, the base repository, and the commits that
// were not scanned.
func printSyncForkSummary(results []checkpoint.ImportResult, report syncForkReport) {
	var synced, present, conflicts, missing, rejected int
	for _, r := range results {
		switch r.Outcome {
		case checkpoint.ImportSynced:
			synced++
		case checkpoint.ImportPresent:
			present++
		case checkpoint.ImportConflict:
			conflicts++
			slog.Warn("checkpoint already exists locally with different content; kept the local copy", "id", r.ID, "detail", r.Reason)
		case checkpoint.ImportMissing:
			missing++
			slog.Warn("checkpoint referenced by the pull request is missing from the fork's checkpoint branch", "id", r.ID)
		case checkpoint.ImportRejected:
			rejected++
			slog.Warn("checkpoint from the fork failed validation and was skipped", "id", r.ID, "reason", r.Reason)
		}
	}
	summary := fmt.Sprintf("synced %d, already present %d, conflicts %d, missing %d", synced, present, conflicts, missing)
	if rejected > 0 {
		summary += fmt.Sprintf(", rejected %d", rejected)
	}
	fmt.Println(summary)

	if report.OriginRepo != "" {
		fmt.Printf("base repository: %s (origin is %s)\n", report.BaseRepo, report.OriginRepo)
	} else {
		fmt.Printf("base repository: %s\n", report.BaseRepo)
	}
	if report.truncated() {
		fmt.Printf("not scanned: commits %d-%d (only the first %d commits of a pull request are listed)\n",
			github.MaxPullCommits+1, report.TotalCommits, github.MaxPullCommits)
	}
}

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
	if pr.Head.Repo == nil || pr.Head.Repo.CloneURL == "" {
		return fmt.Errorf("%s has no head repository (the fork may have been deleted)", ref)
	}
	head := *pr.Head.Repo
	slog.Debug("resolved pull request", "pr", ref.String(), "head_repo", head.FullName, "head_sha", pr.Head.SHA, "base_repo", pr.Base.Repo.FullName)

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

	printSyncForkSummary(results)
	return nil
}

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
// prints the per-outcome counts.
func printSyncForkSummary(results []checkpoint.ImportResult) {
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
}

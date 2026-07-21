package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/git"
)

func newRewindCmd() *cobra.Command {
	var (
		list      bool
		toID      string
		commitSHA string
	)

	cmd := &cobra.Command{
		Use:   "rewind",
		Short: "List or restore checkpoints",
		Long:  `List all captured checkpoints or restore the repository state to a specific checkpoint.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				return runRewindList()
			}
			if toID != "" {
				return runRewindTo(toID)
			}
			if commitSHA != "" {
				return runRewindByCommit(commitSHA)
			}
			return cmd.Help()
		},
	}

	cmd.Flags().BoolVar(&list, "list", false, "list all checkpoints")
	cmd.Flags().StringVar(&toID, "to", "", "restore to a specific checkpoint ID")
	cmd.Flags().StringVar(&commitSHA, "commit", "", "find and restore the checkpoint for the given commit SHA (supports squash-merged commits)")

	return cmd
}

func runRewindList() error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	const branch = "partio/checkpoints/v1"

	// Check branch exists
	_, err = git.ExecGit("rev-parse", "--verify", branch)
	if err != nil {
		return fmt.Errorf("no checkpoint branch found - run 'partio enable' first")
	}

	// List top-level shard directories
	shards, err := git.ExecGit("ls-tree", "--name-only", branch)
	if err != nil || shards == "" {
		fmt.Println("No checkpoints found.")
		return nil
	}

	fmt.Println("Checkpoints:")
	fmt.Println()

	for _, shard := range strings.Split(shards, "\n") {
		// List entries within each shard
		entries, err := git.ExecGit("ls-tree", "--name-only", branch+":"+shard)
		if err != nil {
			continue
		}
		for _, entry := range strings.Split(entries, "\n") {
			cpID := shard + entry
			// Try to read metadata
			metaJSON, err := git.ExecGit("show", branch+":"+shard+"/"+entry+"/metadata.json")
			if err != nil {
				fmt.Printf("  %s  (metadata unavailable)\n", cpID)
				continue
			}

			var meta checkpoint.Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				fmt.Printf("  %s  (invalid metadata)\n", cpID)
				continue
			}

			reachabilityNote := ""
			if !git.CommitReachable(meta.CommitHash) {
				if git.CommitExists(meta.CommitHash) {
					reachabilityNote = "  [squash-merged]"
				} else {
					reachabilityNote = "  [commit pruned]"
				}
			}

			fmt.Printf("  %s  branch=%s  agent=%d%%  created=%s%s\n",
				cpID, meta.Branch, meta.AgentPercent, meta.CreatedAt, reachabilityNote)
		}
	}

	return nil
}

func runRewindTo(id string) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	if len(id) != 12 {
		return fmt.Errorf("checkpoint ID must be 12 characters (got %d)", len(id))
	}

	const branch = "partio/checkpoints/v1"
	shard := id[:2]
	rest := id[2:]

	// Read metadata
	metaJSON, err := git.ExecGit("show", branch+":"+shard+"/"+rest+"/metadata.json")
	if err != nil {
		return fmt.Errorf("checkpoint %s not found", id)
	}

	var meta checkpoint.Metadata
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return fmt.Errorf("invalid checkpoint metadata: %w", err)
	}

	// Read session context
	context, _ := git.ExecGit("show", branch+":"+shard+"/"+rest+"/0/context.md")

	// Resolve the commit hash: if the original was squash-merged it may no
	// longer exist, so fall back to the commit that carries the trailer.
	commitHash := checkpoint.ResolveCommitHash(id, meta.CommitHash)

	fmt.Printf("Rewinding to checkpoint %s\n", id)
	fmt.Printf("  Commit: %s\n", commitHash)
	fmt.Printf("  Branch: %s\n", meta.Branch)
	if context != "" {
		fmt.Printf("  Context:\n%s\n", context)
	}

	// Create a new branch at the resolved commit.
	branchName := fmt.Sprintf("partio/rewind/%s", id)
	_, err = git.ExecGit("checkout", "-b", branchName, commitHash)
	if err != nil {
		return fmt.Errorf("creating rewind branch: %w", err)
	}

	fmt.Printf("  Created branch: %s\n", branchName)
	return nil
}

// runRewindByCommit finds checkpoints associated with the given commit SHA and presents
// them to the user. Supports squash-merged commits by also checking the equivalent
// squash commit in the current branch history when the given SHA is unreachable.
func runRewindByCommit(sha string) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	// Search checkpoints for the given SHA directly
	ids, err := checkpoint.FindByCommit(sha)
	if err != nil {
		return fmt.Errorf("searching checkpoints: %w", err)
	}

	// If not found and the commit exists but is unreachable, check for a squash-merge equivalent
	if len(ids) == 0 && git.CommitExists(sha) && !git.CommitReachable(sha) {
		squash, findErr := git.FindSquashCommit(sha)
		if findErr == nil && squash != "" {
			fmt.Printf("Commit %s is not reachable (squash-merged). Searching checkpoints for equivalent squash commit %s...\n",
				shortSHA(sha), shortSHA(squash))
			ids, err = checkpoint.FindByCommit(squash)
			if err != nil {
				return fmt.Errorf("searching checkpoints for squash commit: %w", err)
			}
		}
	}

	if len(ids) == 0 {
		return fmt.Errorf("no checkpoint found for commit %s", shortSHA(sha))
	}

	if len(ids) == 1 {
		fmt.Printf("Found checkpoint %s for commit %s.\n", ids[0], shortSHA(sha))
		return runRewindTo(ids[0])
	}

	fmt.Printf("Found %d checkpoints for commit %s:\n", len(ids), shortSHA(sha))
	for _, id := range ids {
		fmt.Printf("  %s\n", id)
	}
	fmt.Println("Use 'partio rewind --to <id>' to restore a specific checkpoint.")
	return nil
}

// shortSHA returns the first 8 characters of a SHA for display purposes.
func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/git"
)

func newPruneCmd() *cobra.Command {
	var (
		olderThan string
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete old checkpoints",
		Long:  `Remove checkpoints older than a retention window from the checkpoint branch. The checkpoint linked to the current HEAD is never deleted.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrune(olderThan, dryRun)
		},
	}

	cmd.Flags().StringVar(&olderThan, "older-than", "90d", "retention window (e.g. 30d, 90d)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview what would be deleted without making changes")

	return cmd
}

func parseDuration(s string) (time.Duration, error) {
	if !strings.HasSuffix(s, "d") {
		return 0, fmt.Errorf("unsupported duration format %q (use e.g. 30d, 90d)", s)
	}
	days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
	if err != nil || days <= 0 {
		return 0, fmt.Errorf("invalid duration %q: must be a positive number of days", s)
	}
	return time.Duration(days) * 24 * time.Hour, nil
}

type checkpointEntry struct {
	id        string
	shard     string
	rest      string
	meta      checkpoint.Metadata
	createdAt time.Time
}

func listCheckpoints(branch string) ([]checkpointEntry, error) {
	shards, err := git.ExecGit("ls-tree", "--name-only", branch)
	if err != nil || shards == "" {
		return nil, nil
	}

	var entries []checkpointEntry
	for _, shard := range strings.Split(shards, "\n") {
		if shard == "" {
			continue
		}
		rests, err := git.ExecGit("ls-tree", "--name-only", branch+":"+shard)
		if err != nil {
			continue
		}
		for _, rest := range strings.Split(rests, "\n") {
			if rest == "" {
				continue
			}
			metaJSON, err := git.ExecGit("show", branch+":"+shard+"/"+rest+"/metadata.json")
			if err != nil {
				continue
			}
			var meta checkpoint.Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				continue
			}
			createdAt, err := time.Parse(time.RFC3339, meta.CreatedAt)
			if err != nil {
				continue
			}
			entries = append(entries, checkpointEntry{
				id:        shard + rest,
				shard:     shard,
				rest:      rest,
				meta:      meta,
				createdAt: createdAt,
			})
		}
	}
	return entries, nil
}

func runPrune(olderThan string, dryRun bool) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	retention, err := parseDuration(olderThan)
	if err != nil {
		return err
	}

	const branch = "partio/checkpoints/v1"

	_, err = git.ExecGit("rev-parse", "--verify", branch)
	if err != nil {
		return fmt.Errorf("no checkpoint branch found - run 'partio enable' first")
	}

	headCommit, err := git.ExecGit("rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("getting HEAD: %w", err)
	}

	cutoff := time.Now().Add(-retention)

	all, err := listCheckpoints(branch)
	if err != nil {
		return fmt.Errorf("listing checkpoints: %w", err)
	}
	if len(all) == 0 {
		fmt.Println("No checkpoints found.")
		return nil
	}

	var toPrune []checkpointEntry
	for _, cp := range all {
		if cp.meta.CommitHash == headCommit {
			continue // never prune the checkpoint linked to HEAD
		}
		if cp.createdAt.Before(cutoff) {
			toPrune = append(toPrune, cp)
		}
	}

	if len(toPrune) == 0 {
		fmt.Println("Nothing to prune.")
		return nil
	}

	if dryRun {
		fmt.Printf("Would remove %d of %d checkpoints (dry run):\n", len(toPrune), len(all))
		for _, cp := range toPrune {
			fmt.Printf("  %s  branch=%s  created=%s\n", cp.id, cp.meta.Branch, cp.meta.CreatedAt)
		}
		return nil
	}

	// Build set of entries to prune, keyed by shard → rest
	pruneSet := make(map[string]map[string]bool)
	for _, cp := range toPrune {
		if pruneSet[cp.shard] == nil {
			pruneSet[cp.shard] = make(map[string]bool)
		}
		pruneSet[cp.shard][cp.rest] = true
	}

	// Rebuild the orphan branch tree without pruned entries
	currentTree, err := git.ExecGit("rev-parse", branch+"^{tree}")
	if err != nil {
		return fmt.Errorf("getting current tree: %w", err)
	}

	rootEntries, err := git.ExecGit("ls-tree", currentTree)
	if err != nil {
		return fmt.Errorf("listing root tree: %w", err)
	}

	var newRootLines []string
	for _, line := range strings.Split(rootEntries, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		tabParts := strings.SplitN(line, "\t", 2)
		if len(parts) < 4 || len(tabParts) < 2 {
			continue
		}
		shardName := tabParts[1]

		prunedInShard := pruneSet[shardName]
		if prunedInShard == nil {
			newRootLines = append(newRootLines, line)
			continue
		}

		// Filter entries within this shard
		shardTreeHash := parts[2]
		shardEntries, err := git.ExecGit("ls-tree", shardTreeHash)
		if err != nil {
			newRootLines = append(newRootLines, line)
			continue
		}

		var filteredLines []string
		for _, shardLine := range strings.Split(shardEntries, "\n") {
			if shardLine == "" {
				continue
			}
			shardTabParts := strings.SplitN(shardLine, "\t", 2)
			if len(shardTabParts) < 2 {
				continue
			}
			if prunedInShard[shardTabParts[1]] {
				continue
			}
			filteredLines = append(filteredLines, shardLine)
		}

		if len(filteredLines) == 0 {
			continue // drop empty shard
		}

		newShardHash, err := git.ExecGitStdin(strings.Join(filteredLines, "\n")+"\n", "mktree")
		if err != nil {
			return fmt.Errorf("creating shard tree: %w", err)
		}

		newRootLines = append(newRootLines, fmt.Sprintf("040000 tree %s\t%s", newShardHash, shardName))
	}

	var newRootTree string
	if len(newRootLines) == 0 {
		newRootTree, err = git.ExecGitStdin("\n", "mktree")
	} else {
		newRootTree, err = git.ExecGitStdin(strings.Join(newRootLines, "\n")+"\n", "mktree")
	}
	if err != nil {
		return fmt.Errorf("creating root tree: %w", err)
	}

	parentCommit, err := git.ExecGit("rev-parse", branch)
	if err != nil {
		return fmt.Errorf("getting parent commit: %w", err)
	}

	commitMsg := fmt.Sprintf("prune: removed %d checkpoints older than %s", len(toPrune), olderThan)
	commitHash, err := git.ExecGit("commit-tree", newRootTree, "-p", parentCommit, "-m", commitMsg)
	if err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}

	_, err = git.ExecGit("update-ref", "refs/heads/"+branch, commitHash)
	if err != nil {
		return fmt.Errorf("updating ref: %w", err)
	}

	fmt.Printf("Pruned %d of %d checkpoints (older than %s).\n", len(toPrune), len(all), olderThan)
	return nil
}

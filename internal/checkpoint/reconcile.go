package checkpoint

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// Reconcile fetches the remote checkpoint branch and merges it with the local
// branch so that neither side loses checkpoint data. It is a no-op when:
//   - the repository has no remote named "origin"
//   - the remote does not have the checkpoint branch
//   - local and remote refs are already equal
//
// When only the remote branch exists, Reconcile fast-forwards the local ref.
// When both exist and differ, it builds a union tree (all checkpoints from
// both sides) and creates a merge commit with both tips as parents.
//
// Reconcile errors are returned to the caller; non-critical callers (Write,
// Prune) log them at slog.Warn and continue, while interactive callers
// (partio enable) surface them as visible warnings.
func (s *Store) Reconcile() error {
	// Guard: no remote → nothing to do.
	if _, err := s.git("remote", "get-url", "origin"); err != nil {
		return nil
	}

	// Check whether the remote has the checkpoint branch by querying its refs.
	// ls-remote connects to the remote and lists refs matching the pattern;
	// empty output means the branch is absent (benign), non-nil error means
	// the remote is unreachable (real error).
	lsOut, err := s.git("ls-remote", "origin", "refs/heads/"+checkpointBranch)
	if err != nil {
		return fmt.Errorf("reconcile: ls-remote: %w", err)
	}
	if lsOut == "" {
		// Remote does not have the checkpoint branch — nothing to reconcile.
		return nil
	}

	// Fetch the remote branch into the local remote-tracking ref.
	_, err = s.git("fetch", "--no-tags", "origin",
		"refs/heads/"+checkpointBranch+":refs/remotes/origin/"+checkpointBranch)
	if err != nil {
		return fmt.Errorf("reconcile: fetching remote branch: %w", err)
	}

	// Resolve both local and remote refs.
	localRef, localErr := s.git("rev-parse", "refs/heads/"+checkpointBranch)
	remoteRef, remoteErr := s.git("rev-parse", "refs/remotes/origin/"+checkpointBranch)

	if remoteErr != nil {
		// Fetch succeeded but remote-tracking ref is missing — treat as no-op.
		return nil
	}

	if localErr != nil {
		// Local branch absent, remote present → fast-forward without a merge commit.
		if _, err = s.git("update-ref", "refs/heads/"+checkpointBranch, remoteRef); err != nil {
			return fmt.Errorf("reconcile: fast-forwarding local branch: %w", err)
		}
		return nil
	}

	if localRef == remoteRef {
		return nil
	}

	// Both exist and differ → build a union tree and create a merge commit.
	localEntries, err := s.collectCheckpointEntries(localRef)
	if err != nil {
		return fmt.Errorf("reconcile: collecting local checkpoints: %w", err)
	}

	remoteEntries, err := s.collectCheckpointEntries(remoteRef)
	if err != nil {
		return fmt.Errorf("reconcile: collecting remote checkpoints: %w", err)
	}

	mergedRoot, err := s.buildMergedTree(localEntries, remoteEntries)
	if err != nil {
		return fmt.Errorf("reconcile: building merged tree: %w", err)
	}

	// Merge commit has both tips as parents so git log --graph shows full history.
	commitHash, err := s.commitTreeWithIdentity(mergedRoot, []string{localRef, remoteRef},
		"partio: reconcile checkpoint branches")
	if err != nil {
		return fmt.Errorf("reconcile: creating merge commit: %w", err)
	}

	if _, err = s.git("update-ref", "refs/heads/"+checkpointBranch, commitHash); err != nil {
		return fmt.Errorf("reconcile: updating local ref: %w", err)
	}

	return nil
}

// checkpointEntry holds the location and subtree hash of one checkpoint within
// the shard/rest directory structure on the checkpoint branch.
type checkpointEntry struct {
	shard    string
	rest     string
	treeHash string
}

// collectCheckpointEntries walks the two-level shard/rest structure at tip and
// returns one entry per checkpoint.
func (s *Store) collectCheckpointEntries(tip string) ([]checkpointEntry, error) {
	shards, err := s.git("ls-tree", "--name-only", tip)
	if err != nil {
		return nil, fmt.Errorf("listing shards at %s: %w", tip, err)
	}

	var entries []checkpointEntry
	for _, shard := range strings.Split(shards, "\n") {
		if shard == "" {
			continue
		}
		lines, err := s.git("ls-tree", tip+":"+shard)
		if err != nil {
			// A missing or unreadable shard is non-fatal; skip it.
			continue
		}
		for _, line := range strings.Split(lines, "\n") {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			tabParts := strings.SplitN(line, "\t", 2)
			if len(parts) < 3 || len(tabParts) < 2 {
				continue
			}
			entries = append(entries, checkpointEntry{
				shard:    shard,
				rest:     tabParts[1],
				treeHash: parts[2],
			})
		}
	}

	return entries, nil
}

// buildMergedTree creates a root tree object containing the union of
// checkpoints from both entry lists. When the same checkpoint ID (same rest
// under the same shard) appears on both sides, the local copy is kept —
// checkpoint IDs are cryptographically unique so equal IDs mean equal content.
func (s *Store) buildMergedTree(local, remote []checkpointEntry) (string, error) {
	// merged[shard][rest] = treeHash
	merged := make(map[string]map[string]string)

	add := func(entries []checkpointEntry) {
		for _, e := range entries {
			if merged[e.shard] == nil {
				merged[e.shard] = make(map[string]string)
			}
			// First writer (local) wins; remote fills in only absent IDs.
			if _, exists := merged[e.shard][e.rest]; !exists {
				merged[e.shard][e.rest] = e.treeHash
			}
		}
	}

	add(local)  // local entries take priority
	add(remote) // remote entries fill in anything absent locally

	if len(merged) == 0 {
		return s.mktree(nil)
	}

	shards := make([]string, 0, len(merged))
	for shard := range merged {
		shards = append(shards, shard)
	}
	sort.Strings(shards)

	var rootEntries []treeEntry
	for _, shard := range shards {
		restMap := merged[shard]

		rests := make([]string, 0, len(restMap))
		for rest := range restMap {
			rests = append(rests, rest)
		}
		sort.Strings(rests)

		var shardTreeEntries []treeEntry
		for _, rest := range rests {
			shardTreeEntries = append(shardTreeEntries, treeEntry{
				mode: "040000",
				typ:  "tree",
				hash: restMap[rest],
				name: rest,
			})
		}

		shardTree, err := s.mktree(shardTreeEntries)
		if err != nil {
			return "", fmt.Errorf("creating shard tree %s: %w", shard, err)
		}

		rootEntries = append(rootEntries, treeEntry{
			mode: "040000",
			typ:  "tree",
			hash: shardTree,
			name: shard,
		})
	}

	return s.mktree(rootEntries)
}

// commitTreeWithIdentity creates a git commit object with a fixed committer
// identity so reconciliation commits do not depend on the user's git config.
func (s *Store) commitTreeWithIdentity(tree string, parents []string, msg string) (string, error) {
	args := []string{"commit-tree", tree}
	for _, p := range parents {
		args = append(args, "-p", p)
	}
	args = append(args, "-m", msg)

	cmd := exec.Command("git", args...)
	cmd.Dir = s.repoRoot
	cmd.Env = append(os.Environ(),
		"GIT_COMMITTER_NAME=partio",
		"GIT_COMMITTER_EMAIL=partio@localhost",
		"GIT_AUTHOR_NAME=partio",
		"GIT_AUTHOR_EMAIL=partio@localhost",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

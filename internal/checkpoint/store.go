package checkpoint

import (
	"fmt"
	"os/exec"
	"strings"
)

const checkpointBranch = "partio/checkpoints/v1"

// Store writes checkpoint data to the orphan branch using git plumbing.
type Store struct {
	repoRoot string
}

// NewStore creates a new checkpoint store.
func NewStore(repoRoot string) *Store {
	return &Store{repoRoot: repoRoot}
}

// SessionFiles holds the files to be stored for a session.
type SessionFiles struct {
	ContentHash string
	Context     string
	Diff        string
	FullJSONL   string
	Metadata    SessionMetadata
	Plan        string
	Prompt      string
}

type treeEntry struct {
	mode string
	typ  string
	hash string
	name string
}

func (s *Store) hashObject(content string) (string, error) {
	cmd := exec.Command("git", "hash-object", "-w", "--stdin")
	cmd.Dir = s.repoRoot
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *Store) mktree(entries []treeEntry) (string, error) {
	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s %s %s\t%s", e.mode, e.typ, e.hash, e.name))
	}
	input := strings.Join(lines, "\n") + "\n"

	cmd := exec.Command("git", "mktree")
	cmd.Dir = s.repoRoot
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *Store) getCurrentTree() (string, error) {
	return s.git("rev-parse", checkpointBranch+"^{tree}")
}

func (s *Store) addToTree(currentTree, shard, rest, cpTree string) (string, error) {
	// Read the current root tree
	currentEntries, err := s.git("ls-tree", currentTree)
	if err != nil {
		currentEntries = ""
	}

	// Check if shard directory exists
	var shardTree string
	for _, line := range strings.Split(currentEntries, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 4 && parts[3] == shard {
			shardTree = parts[2]
			break
		}
	}

	// Build shard tree with new entry
	var shardEntries []treeEntry
	if shardTree != "" {
		// Read existing shard entries
		existing, _ := s.git("ls-tree", shardTree)
		for _, line := range strings.Split(existing, "\n") {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				// Parse the tab-separated name
				tabParts := strings.SplitN(line, "\t", 2)
				name := ""
				if len(tabParts) >= 2 {
					name = tabParts[1]
				}
				shardEntries = append(shardEntries, treeEntry{
					mode: parts[0],
					typ:  parts[1],
					hash: parts[2],
					name: name,
				})
			}
		}
	}

	// Add our new checkpoint entry
	shardEntries = append(shardEntries, treeEntry{
		mode: "040000",
		typ:  "tree",
		hash: cpTree,
		name: rest,
	})

	newShardTree, err := s.mktree(shardEntries)
	if err != nil {
		return "", fmt.Errorf("creating shard tree: %w", err)
	}

	// Rebuild root tree
	var rootEntries []treeEntry
	foundShard := false
	for _, line := range strings.Split(currentEntries, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		tabParts := strings.SplitN(line, "\t", 2)
		name := ""
		if len(tabParts) >= 2 {
			name = tabParts[1]
		}

		if len(parts) >= 4 && name == shard {
			// Replace shard entry
			rootEntries = append(rootEntries, treeEntry{
				mode: "040000",
				typ:  "tree",
				hash: newShardTree,
				name: shard,
			})
			foundShard = true
		} else if len(parts) >= 4 {
			rootEntries = append(rootEntries, treeEntry{
				mode: parts[0],
				typ:  parts[1],
				hash: parts[2],
				name: name,
			})
		}
	}

	if !foundShard {
		rootEntries = append(rootEntries, treeEntry{
			mode: "040000",
			typ:  "tree",
			hash: newShardTree,
			name: shard,
		})
	}

	return s.mktree(rootEntries)
}

func (s *Store) git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

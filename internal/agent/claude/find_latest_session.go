package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/partio-io/cli/internal/agent"
)

// FindLatestSession finds the most recently modified JSONL session file.
func (d *Detector) FindLatestSession(repoRoot string) (string, *agent.SessionData, error) {
	sessionDir, err := d.FindSessionDir(repoRoot)
	if err != nil {
		return "", nil, err
	}

	// Find all .jsonl files
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return "", nil, fmt.Errorf("reading session directory: %w", err)
	}

	var jsonlFiles []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			jsonlFiles = append(jsonlFiles, e)
		}
	}

	if len(jsonlFiles) == 0 {
		return "", nil, fmt.Errorf("no JSONL session files found in %s", sessionDir)
	}

	// Sort by modification time (most recent first)
	sort.Slice(jsonlFiles, func(i, j int) bool {
		ii, _ := jsonlFiles[i].Info()
		jj, _ := jsonlFiles[j].Info()
		return ii.ModTime().After(jj.ModTime())
	})

	latestPath := filepath.Join(sessionDir, jsonlFiles[0].Name())

	// Parse the JSONL file
	data, err := ParseJSONL(latestPath)
	if err != nil {
		return latestPath, nil, fmt.Errorf("parsing JSONL: %w", err)
	}

	return latestPath, data, nil
}

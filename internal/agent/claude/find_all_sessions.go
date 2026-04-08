package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/partio-io/cli/internal/agent"
)

// SubAgentSession holds a subagent session path and its parsed data.
type SubAgentSession struct {
	Path string
	Data *agent.SessionData
}

// FindAllSessions returns all sessions in the session directory, classifying
// the one with the earliest start time as the primary session and the rest as
// subagent sessions. Subagent sessions are spawned by the primary agent during
// the session and have a later start time.
func (d *Detector) FindAllSessions(repoRoot string) (string, *agent.SessionData, []SubAgentSession, error) {
	sessionDir, err := d.FindSessionDir(repoRoot)
	if err != nil {
		return "", nil, nil, err
	}

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return "", nil, nil, fmt.Errorf("reading session directory: %w", err)
	}

	var jsonlPaths []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			jsonlPaths = append(jsonlPaths, filepath.Join(sessionDir, e.Name()))
		}
	}

	if len(jsonlPaths) == 0 {
		return "", nil, nil, fmt.Errorf("no JSONL session files found in %s", sessionDir)
	}

	type sessionCandidate struct {
		path      string
		data      *agent.SessionData
		startTime time.Time
	}

	var candidates []sessionCandidate
	for _, p := range jsonlPaths {
		data, err := ParseJSONL(p)
		if err != nil {
			continue
		}
		var startTime time.Time
		for _, msg := range data.Transcript {
			if !msg.Timestamp.IsZero() {
				startTime = msg.Timestamp
				break
			}
		}
		// Fall back to file modification time if no message timestamps available.
		if startTime.IsZero() {
			if info, err := os.Stat(p); err == nil {
				startTime = info.ModTime()
			}
		}
		candidates = append(candidates, sessionCandidate{path: p, data: data, startTime: startTime})
	}

	if len(candidates) == 0 {
		return "", nil, nil, fmt.Errorf("no parseable JSONL session files found in %s", sessionDir)
	}

	// Sort ascending by start time: the earliest session is the primary.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].startTime.IsZero() {
			return false
		}
		if candidates[j].startTime.IsZero() {
			return true
		}
		return candidates[i].startTime.Before(candidates[j].startTime)
	})

	primary := candidates[0]
	var subAgents []SubAgentSession
	for _, c := range candidates[1:] {
		subAgents = append(subAgents, SubAgentSession{Path: c.path, Data: c.data})
	}

	return primary.path, primary.data, subAgents, nil
}

package checksverdict

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/auditgate"
)

// Write saves the verdict where the gate expects to find it, creating
// the directory if the check runs before anything else has.
func Write(path string, v auditgate.Verdict) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create verdict directory: %w", err)
		}
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode verdict: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write verdict: %w", err)
	}
	return nil
}

package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// HookManager represents an external Git hook manager that may conflict with partio.
type HookManager struct {
	Name   string
	Reason string
}

// DetectExternalHookManagers checks the repository root for signs of external
// Git hook managers (Husky, Lefthook, Overcommit) and returns any that are found.
func DetectExternalHookManagers(repoRoot string) []HookManager {
	var found []HookManager

	// Husky: .husky/ directory
	if info, err := os.Stat(filepath.Join(repoRoot, ".husky")); err == nil && info.IsDir() {
		found = append(found, HookManager{Name: "Husky", Reason: ".husky/ directory found"})
	}

	// Husky: prepare script in package.json
	if !hasHusky(found) {
		if detectHuskyInPackageJSON(repoRoot) {
			found = append(found, HookManager{Name: "Husky", Reason: "\"prepare\" script found in package.json"})
		}
	}

	// Lefthook: lefthook.yml or .lefthook.yml
	if _, err := os.Stat(filepath.Join(repoRoot, "lefthook.yml")); err == nil {
		found = append(found, HookManager{Name: "Lefthook", Reason: "lefthook.yml found"})
	} else if _, err := os.Stat(filepath.Join(repoRoot, ".lefthook.yml")); err == nil {
		found = append(found, HookManager{Name: "Lefthook", Reason: ".lefthook.yml found"})
	}

	// Overcommit: .overcommit.yml
	if _, err := os.Stat(filepath.Join(repoRoot, ".overcommit.yml")); err == nil {
		found = append(found, HookManager{Name: "Overcommit", Reason: ".overcommit.yml found"})
	}

	return found
}

func hasHusky(managers []HookManager) bool {
	for _, m := range managers {
		if m.Name == "Husky" {
			return true
		}
	}
	return false
}

func detectHuskyInPackageJSON(repoRoot string) bool {
	data, err := os.ReadFile(filepath.Join(repoRoot, "package.json"))
	if err != nil {
		return false
	}

	var pkg struct {
		Scripts struct {
			Prepare string `json:"prepare"`
		} `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	return pkg.Scripts.Prepare != ""
}

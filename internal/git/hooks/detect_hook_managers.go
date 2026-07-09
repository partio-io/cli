package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// HookManager represents a detected external git hook manager.
type HookManager struct {
	Name         string
	Instructions string
}

// DetectHookManagers checks the repository root for known external git hook managers.
// It returns a slice of detected managers with integration instructions.
// Detected managers: Husky, Lefthook, Overcommit.
func DetectHookManagers(repoRoot string) []HookManager {
	var detected []HookManager

	if detectHusky(repoRoot) {
		detected = append(detected, HookManager{
			Name: "Husky",
			Instructions: "Add partio hook calls to your Husky hook scripts. For example, in .husky/pre-commit add:\n" +
				"    partio _hook pre-commit\n" +
				"  and in .husky/post-commit add:\n" +
				"    partio _hook post-commit",
		})
	}

	if detectLefthook(repoRoot) {
		detected = append(detected, HookManager{
			Name: "Lefthook",
			Instructions: "Add partio hook calls to your lefthook.yml. For example:\n" +
				"  pre-commit:\n" +
				"    commands:\n" +
				"      partio:\n" +
				"        run: partio _hook pre-commit\n" +
				"  post-commit:\n" +
				"    commands:\n" +
				"      partio:\n" +
				"        run: partio _hook post-commit",
		})
	}

	if detectOvercommit(repoRoot) {
		detected = append(detected, HookManager{
			Name: "Overcommit",
			Instructions: "Add partio hook calls to your .overcommit.yml. For example:\n" +
				"  PreCommit:\n" +
				"    ExecuteScript:\n" +
				"      partio:\n" +
				"        command: ['partio', '_hook', 'pre-commit']\n" +
				"  PostCommit:\n" +
				"    ExecuteScript:\n" +
				"      partio:\n" +
				"        command: ['partio', '_hook', 'post-commit']",
		})
	}

	return detected
}

func detectHusky(repoRoot string) bool {
	// Check for .husky/ directory
	if info, err := os.Stat(filepath.Join(repoRoot, ".husky")); err == nil && info.IsDir() {
		return true
	}

	// Check for "husky" key in package.json
	data, err := os.ReadFile(filepath.Join(repoRoot, "package.json"))
	if err != nil {
		return false
	}
	var pkg map[string]json.RawMessage
	if json.Unmarshal(data, &pkg) != nil {
		return false
	}
	// husky can appear in devDependencies, dependencies, or as a top-level "husky" config key
	for _, key := range []string{"husky", "devDependencies", "dependencies"} {
		raw, ok := pkg[key]
		if !ok {
			continue
		}
		if key == "husky" {
			return true
		}
		var deps map[string]json.RawMessage
		if json.Unmarshal(raw, &deps) == nil {
			if _, found := deps["husky"]; found {
				return true
			}
		}
	}
	return false
}

func detectLefthook(repoRoot string) bool {
	for _, name := range []string{"lefthook.yml", "lefthook.yaml", ".lefthook.yml", ".lefthook.yaml"} {
		if _, err := os.Stat(filepath.Join(repoRoot, name)); err == nil {
			return true
		}
	}
	return false
}

func detectOvercommit(repoRoot string) bool {
	_, err := os.Stat(filepath.Join(repoRoot, ".overcommit.yml"))
	return err == nil
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/config"
)

// stealthOverride is one repository-level setting that defeats stealth.
type stealthOverride struct {
	file  string // path relative to the repo root, e.g. ".partio/settings.json"
	key   string // dotted key, e.g. "strategy_options.push_sessions"
	value bool   // the value the file sets
}

func (o stealthOverride) String() string {
	return fmt.Sprintf("%s sets %s=%t", o.file, o.key, o.value)
}

// stealthOverrides reports every repository-level setting that outranks the
// global stealth settings: strategy_options.push_sessions=true or
// strategy_options.disable_trailers=false. Layers are checked from highest to
// lowest priority. A missing or malformed file or value is skipped, the same
// way config.Load treats it.
func stealthOverrides(repoRoot string) []stealthOverride {
	var found []stealthOverride
	for _, name := range []string{"settings.local.json", "settings.json"} {
		rel := filepath.Join(config.PartioDir, name)
		opts := readStrategyOptions(filepath.Join(repoRoot, rel))
		if opts == nil {
			continue
		}
		if v, ok := boolOption(opts, "push_sessions"); ok && v {
			found = append(found, stealthOverride{file: rel, key: "strategy_options.push_sessions", value: true})
		}
		if v, ok := boolOption(opts, "disable_trailers"); ok && !v {
			found = append(found, stealthOverride{file: rel, key: "strategy_options.disable_trailers", value: false})
		}
	}
	return found
}

// readStrategyOptions returns the strategy_options object of the settings file
// at path, or nil if the file or the object is missing or malformed.
func readStrategyOptions(path string) map[string]json.RawMessage {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	var opts map[string]json.RawMessage
	if err := json.Unmarshal(raw["strategy_options"], &opts); err != nil {
		return nil
	}
	return opts
}

func boolOption(opts map[string]json.RawMessage, key string) (value, ok bool) {
	v, present := opts[key]
	if !present {
		return false, false
	}
	if err := json.Unmarshal(v, &value); err != nil {
		return false, false
	}
	return value, true
}

// warnStealthOverrides reloads the effective config after the global stealth
// write and warns when a repository setting keeps stealth from taking effect.
// It never edits repository files to force stealth.
func warnStealthOverrides(repoRoot string) {
	effective, err := config.Load(repoRoot)
	if err == nil && stealthMode(effective) {
		return
	}

	overrides := stealthOverrides(repoRoot)
	if len(overrides) == 0 {
		fmt.Println("\nWarning: stealth is not in effect in this repo; the effective config does not have push_sessions=false and disable_trailers=true.")
		return
	}
	for _, o := range overrides {
		fmt.Printf("\nWarning: %s; stealth is not in effect in this repo.\n", o)
	}
	fmt.Println("  Repository settings outrank the global config. partio does not change them; edit them yourself if you want stealth here.")
}

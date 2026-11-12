package main

import "github.com/partio-io/cli/internal/config"

// stealthModeLine is the line `partio status` prints when stealth is in effect.
const stealthModeLine = "Mode:       stealth (checkpoints local only; not pushed, no commit trailers)"

// stealthMode reports whether the effective config is in stealth mode: the
// checkpoint branch is never pushed and commits get no Partio trailers.
// Stealth is derived from these two options rather than stored as its own key.
func stealthMode(cfg config.Config) bool {
	return !cfg.StrategyOptions.PushSessions && cfg.StrategyOptions.DisableTrailers
}

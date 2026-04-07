package agent

import (
	"fmt"
)

// NewDetectorFunc is a factory function that creates a Detector.
type NewDetectorFunc func() Detector

// registry maps agent names to their factory functions.
var registry = map[string]NewDetectorFunc{}

// Register adds a detector factory to the registry.
func Register(name string, fn NewDetectorFunc) {
	registry[name] = fn
}

// NewDetector returns a Detector for the given agent name.
// Returns an error if no detector is registered for that name.
func NewDetector(name string) (Detector, error) {
	fn, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %s", name)
	}
	return fn(), nil
}

package agent

import "log/slog"

// Detect iterates over the provided detectors and returns the first one
// whose process is currently running. If no agent is detected, it returns nil.
func Detect(detectors []Detector) Detector {
	for _, d := range detectors {
		running, err := d.IsRunning()
		if err != nil {
			slog.Debug("error checking agent", "agent", d.Name(), "error", err)
			continue
		}
		if running {
			return d
		}
	}
	return nil
}

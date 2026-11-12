package main

import (
	"testing"

	"github.com/partio-io/cli/internal/config"
)

func TestStealthMode(t *testing.T) {
	tests := []struct {
		name            string
		pushSessions    bool
		disableTrailers bool
		want            bool
	}{
		{name: "push on, trailers on", pushSessions: true, disableTrailers: false, want: false},
		{name: "push on, trailers off", pushSessions: true, disableTrailers: true, want: false},
		{name: "push off, trailers on", pushSessions: false, disableTrailers: false, want: false},
		{name: "push off, trailers off", pushSessions: false, disableTrailers: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Defaults()
			cfg.StrategyOptions.PushSessions = tt.pushSessions
			cfg.StrategyOptions.DisableTrailers = tt.disableTrailers

			if got := stealthMode(cfg); got != tt.want {
				t.Errorf("stealthMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

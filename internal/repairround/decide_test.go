package repairround

import "testing"

func TestDecide(t *testing.T) {
	const check = "dead-code"
	repair := Subject(check, 41)
	other := Subject("lint-test", 41)

	tests := []struct {
		name      string
		subjects  []string
		maxRounds int
		want      Decision
	}{
		{
			name:      "empty branch",
			subjects:  nil,
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: true, Round: 1, Prior: 0},
		},
		{
			name:      "no prior rounds",
			subjects:  []string{"feat: add the thing", "test: cover it"},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: true, Round: 1, Prior: 0},
		},
		{
			name:      "one prior round",
			subjects:  []string{"feat: add the thing", repair},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: true, Round: 2, Prior: 1},
		},
		{
			name:      "at the cap",
			subjects:  []string{"feat: add the thing", repair, repair, repair},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: false, Round: 4, Prior: 3},
		},
		{
			name:      "over the cap",
			subjects:  []string{repair, repair, repair, repair},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: false, Round: 5, Prior: 4},
		},
		{
			name:      "another check's rounds interleaved",
			subjects:  []string{other, repair, other, other},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: true, Round: 2, Prior: 1},
		},
		{
			name:      "another check at its own cap",
			subjects:  []string{other, other, other},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: true, Round: 1, Prior: 0},
		},
		{
			name: "human commits interleaved",
			subjects: []string{
				"feat: add the thing",
				repair,
				"fix: address review",
				"docs: mention audit(dead-code): in the runbook",
				"revert: audit(dead-code): fix findings (#41)",
			},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: true, Round: 2, Prior: 1},
		},
		{
			name:      "near-miss subject only",
			subjects:  []string{"audit: fix dead-code findings (#620)"},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: true, Round: 1, Prior: 0},
		},
		{
			name:      "the last allowed round",
			subjects:  []string{repair, repair},
			maxRounds: DefaultMaxRounds,
			want:      Decision{Allowed: true, Round: 3, Prior: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decide(tt.subjects, check, tt.maxRounds)
			if got != tt.want {
				t.Errorf("Decide = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// The cap is three rounds per check, so the third attempt is the last
// one allowed and the fourth is declined.
func TestDefaultMaxRoundsIsThree(t *testing.T) {
	if DefaultMaxRounds != 3 {
		t.Fatalf("DefaultMaxRounds = %d, want 3", DefaultMaxRounds)
	}
	repair := Subject("dead-code", 41)
	var branch []string
	for round := 1; round <= DefaultMaxRounds; round++ {
		d := Decide(branch, "dead-code", DefaultMaxRounds)
		if !d.Allowed || d.Round != round {
			t.Fatalf("round %d: Decide = %+v, want allowed round %d", round, d, round)
		}
		branch = append(branch, repair)
	}
	if d := Decide(branch, "dead-code", DefaultMaxRounds); d.Allowed {
		t.Errorf("after %d rounds: Decide = %+v, want declined", DefaultMaxRounds, d)
	}
}

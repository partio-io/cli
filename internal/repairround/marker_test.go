package repairround

import "testing"

func TestSubjectNamesItsCheck(t *testing.T) {
	got := Subject("dead-code", 622)
	want := "audit(dead-code): fix findings (#622)"
	if got != want {
		t.Errorf("Subject = %q, want %q", got, want)
	}
	check, ok := CheckOf(got)
	if !ok || check != "dead-code" {
		t.Errorf("CheckOf(%q) = %q, %v; want \"dead-code\", true", got, check, ok)
	}
}

func TestCheckOf(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		want    string
		wantOK  bool
	}{
		{"marker", "audit(dead-code): fix findings (#622)", "dead-code", true},
		{"other check", "audit(lint-test): fix findings (#622)", "lint-test", true},
		{"digits in the name", "audit(e2e-need): fix findings (#622)", "e2e-need", true},

		// Near misses. Each of these must count for nothing.
		{"unqualified predecessor", "audit: fix dead-code findings (#620)", "", false},
		{"human scope", "fix(audit): tighten the loop guard", "", false},
		{"quoted mid-subject", "docs: explain audit(dead-code): markers", "", false},
		{"no space after colon", "audit(dead-code):fix findings", "", false},
		{"no colon", "audit(dead-code) fix findings", "", false},
		{"empty check", "audit(): fix findings (#622)", "", false},
		{"unclosed check", "audit(dead-code fix findings", "", false},
		{"capitalized check", "audit(Dead-Code): fix findings (#622)", "", false},
		{"spaced check", "audit(dead code): fix findings (#622)", "", false},
		{"leading whitespace", " audit(dead-code): fix findings (#622)", "", false},
		{"ordinary commit", "feat: add the thing", "", false},
		{"empty subject", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CheckOf(tt.subject)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("CheckOf(%q) = %q, %v; want %q, %v", tt.subject, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

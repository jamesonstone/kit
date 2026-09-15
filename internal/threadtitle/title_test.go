package threadtitle

import (
	"strings"
	"testing"
)

func TestResolveOwnership(t *testing.T) {
	tests := []struct {
		name, current, objective, reason string
		accurate, changed                bool
	}{
		{"initial", "", "thread naming", "initial_title", false, true},
		{"evolved work", "[kit] agents / initialization", "thread naming", "ownership_changed", false, true},
		{"verification stage", "[kit] agents / thread naming", "verify thread naming", "ownership_unchanged", true, false},
		{"fork new owner", "[kit] agents / thread naming", "host adapters", "ownership_changed", false, true},
		{"identical", "[kit] agents / thread naming", "thread naming", "cosmetic_only", false, false},
		{"capitalization punctuation whitespace", " [KIT]  Agents: / Thread   Naming! ", "thread naming", "cosmetic_only", false, false},
		{"accurate without current", " ", "thread naming", "initial_title", true, true},
		{"token boundaries", "[kit] agents / threadnaming", "thread naming", "ownership_changed", false, true},
		{"acronym boundaries", "[kit] agents / J.W.T validation", "JWT validation", "ownership_changed", false, true},
		{"language symbols", "[kit] agents / C++ support", "C support", "ownership_changed", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(Input{Scope: "kit", Domain: "agents", Objective: tt.objective, Current: tt.current, Accurate: tt.accurate})
			if err != nil {
				t.Fatal(err)
			}
			if got.Changed != tt.changed || got.Reason != tt.reason {
				t.Fatalf("unexpected decision: %+v", got)
			}
			want := tt.current
			if tt.changed {
				want = "[kit] agents / " + tt.objective
			}
			if got.Title != want {
				t.Fatalf("title = %q, want %q", got.Title, want)
			}
		})
	}
}

func TestResolveNormalizesWithoutTruncating(t *testing.T) {
	objective := "preserve meaningful outcomes " + strings.Repeat("across integrations ", 8)
	got, err := Resolve(Input{Scope: "  release-program  ", Domain: "host\t adapters", Objective: objective})
	if err != nil {
		t.Fatal(err)
	}
	want := "[release-program] host adapters / " + strings.TrimSpace(objective)
	if got.Title != want {
		t.Fatalf("title = %q, want %q", got.Title, want)
	}
}

func TestResolveRejectsInvalidComponents(t *testing.T) {
	for _, field := range []string{"scope", "domain", "objective"} {
		for _, value := range []string{"", " \t", "a\nb", "a\rb", "a\u2028b", "a\u2029b", "a\x00b", "a/b", "a[b", "a]b"} {
			t.Run(field+"/"+value, func(t *testing.T) {
				in := Input{Scope: "kit", Domain: "agents", Objective: "thread naming", Current: "keep", Accurate: true}
				switch field {
				case "scope":
					in.Scope = value
				case "domain":
					in.Domain = value
				case "objective":
					in.Objective = value
				}
				if _, err := Resolve(in); err == nil {
					t.Fatalf("accepted invalid %s %q", field, value)
				}
			})
		}
	}
	for _, scope := range []string{"kit+labcore", "kit, labcore"} {
		if _, err := Resolve(Input{Scope: scope, Domain: "agents", Objective: "thread naming"}); err == nil {
			t.Fatalf("accepted repository list %q", scope)
		}
	}
}

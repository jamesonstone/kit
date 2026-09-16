package threadtitle

import "testing"

// These cases verify formatting/decisions after the active agent has resolved
// meaning. Natural-language derivation is evaluated separately in the spec.
func TestUserOwnershipScenarios(t *testing.T) {
	tests := []struct {
		name, scope, domain, objective, current, want string
		accurate, changed                             bool
	}{
		{"initial webhook 403", "labcore", "eventing", "prod webhook 403", "", "[labcore] eventing / prod webhook 403", false, true},
		{"authorization root cause", "labcore", "eventing", "prod JWT routing", "[labcore] eventing / prod webhook 403", "[labcore] eventing / prod JWT routing", false, true},
		{"remaining delivery verification", "labcore", "eventing", "prod delivery verification", "[labcore] eventing / prod JWT routing", "[labcore] eventing / prod delivery verification", false, true},
		{"edit tests deploy logs", "labcore", "eventing", "prod delivery verification", "[labcore] eventing / prod delivery verification", "[labcore] eventing / prod delivery verification", true, false},
		{"EUID child", "labcore", "accessioning", "canonical EUID", "[labcore] accessioning / barcode identity", "[labcore] accessioning / canonical EUID", false, true},
		{"UI child", "labcore-ui", "accessioning", "partner barcode UX", "[labcore] accessioning / barcode identity", "[labcore-ui] accessioning / partner barcode UX", false, true},
		{"existing program scope", "production-line", "accessioning", "barcode identity", "", "[production-line] accessioning / barcode identity", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(Input{Scope: tt.scope, Domain: tt.domain, Objective: tt.objective, Current: tt.current, Accurate: tt.accurate})
			if err != nil {
				t.Fatal(err)
			}
			if got.Title != tt.want || got.Changed != tt.changed {
				t.Fatalf("got %+v, want title %q changed %v", got, tt.want, tt.changed)
			}
		})
	}
}

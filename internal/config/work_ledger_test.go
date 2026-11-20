package config

import "testing"

func TestWorkLedgerConfiguration(t *testing.T) {
	valid := WorkLedgerConfig{Provider: "plane", Plane: PlaneLedgerConfig{Workspace: "reasoning-machines", Project: "WL", Name: "Work Ledger"}}
	for _, tc := range []struct {
		name    string
		cfg     WorkLedgerConfig
		enabled bool
		invalid bool
	}{
		{"omitted", WorkLedgerConfig{}, false, false},
		{"none", WorkLedgerConfig{Provider: "none"}, false, false},
		{"plane", valid, true, false},
		{"missing workspace", WorkLedgerConfig{Provider: "plane", Plane: PlaneLedgerConfig{Project: "WL"}}, true, true},
		{"URL workspace", WorkLedgerConfig{Provider: "plane", Plane: PlaneLedgerConfig{Workspace: "https://app.plane.so/reasoning-machines", Project: "WL"}}, true, true},
		{"title as identifier", WorkLedgerConfig{Provider: "plane", Plane: PlaneLedgerConfig{Workspace: "reasoning-machines", Project: "Work Ledger"}}, true, true},
		{"reserved linear", WorkLedgerConfig{Provider: "linear"}, true, true},
		{"reserved jira", WorkLedgerConfig{Provider: "jira"}, true, true},
		{"reserved notion", WorkLedgerConfig{Provider: "notion"}, true, true},
		{"unknown", WorkLedgerConfig{Provider: "trello"}, true, true},
		{"bad state", WorkLedgerConfig{Provider: "plane", Plane: PlaneLedgerConfig{Workspace: "reasoning-machines", Project: "WL", States: PlaneLedgerStates{Active: "Active\nDone"}}}, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cfg.Enabled() != tc.enabled {
				t.Fatalf("Enabled() = %v", tc.cfg.Enabled())
			}
			if (tc.cfg.Validate() != nil) != tc.invalid {
				t.Fatalf("Validate() = %v", tc.cfg.Validate())
			}
		})
	}
	states := valid.Plane.States.Resolved()
	if states.Active != "Active" || states.InReview != "In review" || states.Done != "Done" {
		t.Fatalf("defaults = %#v", states)
	}
	custom := PlaneLedgerStates{Active: "Doing", InReview: "Review", Done: "Finished"}.Resolved()
	if custom.Active != "Doing" || custom.InReview != "Review" || custom.Done != "Finished" {
		t.Fatalf("custom = %#v", custom)
	}
}

func TestWorkLedgerRoundTrip(t *testing.T) {
	root := t.TempDir()
	cfg := Default()
	cfg.WorkLedger = WorkLedgerConfig{Provider: "plane", Plane: PlaneLedgerConfig{Workspace: "reasoning-machines", Project: "WL", Name: "Work Ledger", States: PlaneLedgerStates{Active: "Active"}}}
	if err := Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, inspection, err := LoadWithInspection(root)
	if err != nil || inspection.HasErrors() || loaded.WorkLedger != cfg.WorkLedger {
		t.Fatalf("round trip: %v %v %#v", err, inspection, loaded)
	}
	if err := loaded.WorkLedger.Validate(); err != nil {
		t.Fatal(err)
	}
	loaded.WorkLedger.Provider = "plane"
	loaded.WorkLedger.Plane.Workspace = ""
	if err := Save(root, loaded); err != nil {
		t.Fatal(err)
	}
	_, inspection, err = LoadWithInspection(root)
	if err != nil || !inspection.HasErrors() {
		t.Fatalf("missing workspace must be reported: %v %v", err, inspection)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("Load accepted an invalid work ledger")
	}
}

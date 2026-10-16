package config

import "testing"

func TestIssueProviderConfiguration(t *testing.T) {
	valid := IssueConfig{Provider: "linear", Linear: LinearConfig{Workspace: "acme", TeamID: "11111111-1111-4111-8111-111111111111", ProjectID: "22222222-2222-4222-8222-222222222222"}}
	for _, tc := range []struct {
		name    string
		cfg     IssueConfig
		want    string
		invalid bool
	}{
		{"legacy", IssueConfig{}, "github", false},
		{"github", IssueConfig{Provider: "github"}, "github", false},
		{"linear", valid, "linear", false},
		{"unknown", IssueConfig{Provider: "linera"}, "linera", true},
		{"missing target", IssueConfig{Provider: "linear"}, "linear", true},
		{"ambiguous team name", IssueConfig{Provider: "linear", Linear: LinearConfig{Workspace: "acme", TeamID: "Engineering"}}, "linear", true},
		{"bad project", IssueConfig{Provider: "linear", Linear: LinearConfig{Workspace: "acme", TeamID: valid.Linear.TeamID, ProjectID: "Website"}}, "linear", true},
		{"URL not slug", IssueConfig{Provider: "linear", Linear: LinearConfig{Workspace: "https://linear.app/acme", TeamID: valid.Linear.TeamID}}, "linear", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cfg.EffectiveProvider() != tc.want {
				t.Fatal(tc.cfg)
			}
			if (tc.cfg.Validate() != nil) != tc.invalid {
				t.Fatalf("Validate() = %v", tc.cfg.Validate())
			}
		})
	}
	root := t.TempDir()
	cfg := Default()
	cfg.Issues = valid
	if err := Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, inspection, err := LoadWithInspection(root)
	if err != nil || inspection.HasErrors() || loaded.Issues != valid {
		t.Fatalf("round trip: %v %v %#v", err, inspection, loaded)
	}
	loaded.Issues.Provider = "github"
	if err := Save(root, loaded); err != nil {
		t.Fatal(err)
	}
	switched, err := Load(root)
	if err != nil || switched.Issues.EffectiveProvider() != "github" || switched.Issues.Linear != valid.Linear {
		t.Fatalf("switch: %v %#v", err, switched)
	}
	loaded.Issues.Provider = "linear"
	loaded.Issues.Linear.TeamID = ""
	if err := Save(root, loaded); err != nil {
		t.Fatal(err)
	}
	_, inspection, err = LoadWithInspection(root)
	if err != nil || !inspection.HasErrors() {
		t.Fatalf("missing target must be reported: %v %v", err, inspection)
	}
}

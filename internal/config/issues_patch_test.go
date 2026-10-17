package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPlanIssueConfigPreservesUnrelatedYAML(t *testing.T) {
	before := []byte("# project comment\nschema_version: 2\ncustom: # keep comment\n  nested: [one, two]\naws:\n  enabled: false\nissues:\n  extension: keep\n  provider: github\n")
	linear := IssueConfig{Provider: "linear", Linear: LinearConfig{Workspace: "acme", TeamID: "11111111-1111-4111-8111-111111111111", ProjectID: "22222222-2222-4222-8222-222222222222"}}
	current := before
	for _, next := range []IssueConfig{linear, {Provider: "github", Linear: linear.Linear}, linear} {
		after, err := PlanIssueConfig(current, next)
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := yaml.Unmarshal(after, &doc); err != nil {
			t.Fatal(err)
		}
		if doc["custom"] == nil || doc["aws"] == nil || !strings.Contains(string(after), "# keep comment") || !strings.Contains(string(after), "extension: keep") {
			t.Fatalf("unrelated content lost: %s", after)
		}
		repeated, err := PlanIssueConfig(after, next)
		if err != nil || !bytes.Equal(after, repeated) {
			t.Fatalf("not repeatable: %v", err)
		}
		current = after
	}
	linear.Linear.ProjectID = ""
	after, err := PlanIssueConfig(current, linear)
	if err != nil || strings.Contains(string(after), "project_id") {
		t.Fatalf("clear project: %s %v", after, err)
	}
	if got, err := PlanIssueConfig(before, IssueConfig{Provider: "linear"}); err == nil || got != nil {
		t.Fatal("invalid target accepted")
	}
}

func TestApplyIssueConfigAtomicPreconditions(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ConfigFileName)
	before := []byte("schema_version: 2\n")
	after := []byte("schema_version: 2\nissues: {provider: github}\n")
	if err := os.WriteFile(path, before, 0600); err != nil {
		t.Fatal(err)
	}
	if err := ApplyIssueConfig(path, []byte("stale"), after); err == nil {
		t.Fatal("stale write accepted")
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, before) {
		t.Fatal("failed write changed config")
	}
	if err := ApplyIssueConfig(path, before, after); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal("permissions changed")
	}
	if err := ApplyIssueConfig(path, after, after); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if err := ApplyIssueConfig(link, after, before); err == nil {
		t.Fatal("symlink write accepted")
	}
	got, _ = os.ReadFile(path)
	if !bytes.Equal(got, after) {
		t.Fatal("symlink target changed")
	}
}

func TestIssuePatchRejectsAliasedAndAnchoredTargets(t *testing.T) {
	for _, before := range []string{
		"defaults: &defaults\n  issues: {provider: github, extension: keep}\n<<: *defaults\n",
		"defaults: &tracker {provider: github, extension: keep}\nissues: *tracker\n",
		"issues: &tracker {provider: github, extension: keep}\nother: *tracker\n",
		"issues:\n  provider: github\n  linear: &target {workspace: acme}\nother: *target\n",
	} {
		desired := IssueConfig{Provider: "linear", Linear: LinearConfig{Workspace: "acme", TeamID: "11111111-1111-4111-8111-111111111111"}}
		got, err := PlanIssueConfig([]byte(before), desired)
		if err == nil || got != nil || !strings.Contains(err.Error(), "manual review") {
			t.Fatalf("alias accepted: %s %v", got, err)
		}
	}
	// An unrelated alias is safe and remains untouched.
	before := []byte("custom: &data {keep: yes}\nother: *data\nissues: {provider: linear, linear: {workspace: acme, team_id: 11111111-1111-4111-8111-111111111111}}\n")
	got, err := PlanIssueConfig(before, IssueConfig{Provider: "github"})
	if err != nil || !strings.Contains(string(got), "*data") {
		t.Fatalf("unrelated alias: %s %v", got, err)
	}
}

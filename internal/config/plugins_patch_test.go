package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfigOmitsPlugins(t *testing.T) {
	data, err := yaml.Marshal(Default())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "plugins") {
		t.Fatalf("default config contains plugins:\n%s", data)
	}
}

func TestPlanPluginEnabledPreservesUnrelatedYAML(t *testing.T) {
	before := []byte("# project comment\nschema_version: 2\nissues:\n  provider: linear\n  linear:\n    workspace: acme\n    team_id: 11111111-1111-4111-8111-111111111111\ncustom: keep\n")
	enabled, err := PlanPluginEnabled(before, []string{"layer2context"})
	if err != nil {
		t.Fatal(err)
	}
	text := string(enabled)
	for _, want := range []string{"# project comment", "workspace: acme", "custom: keep", "layer2context"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s\n%s", want, text)
		}
	}
	again, err := PlanPluginEnabled(enabled, []string{"layer2context"})
	if err != nil || string(again) != text {
		t.Fatal("enable was not idempotent")
	}
	disabled, err := PlanPluginEnabled(enabled, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(disabled), "plugins") || !strings.Contains(string(disabled), "workspace: acme") || !strings.Contains(string(disabled), "# project comment") {
		t.Fatalf("disable:\n%s", disabled)
	}
}

func TestPlanPluginEnabledRejectsBadIDs(t *testing.T) {
	before := []byte("schema_version: 2\n")
	for _, id := range []string{"", "Bad", "-no", "no-", "a--b", "1abc"} {
		if _, err := PlanPluginEnabled(before, []string{id}); err == nil {
			t.Fatalf("accepted %q", id)
		}
	}
}

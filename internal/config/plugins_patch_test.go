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

func TestPlanPluginEnabledOverridesMergedList(t *testing.T) {
	before := []byte("schema_version: 2\nplugins:\n  <<:\n    enabled: [fixture]\n  note: keep\n")
	disabled, err := PlanPluginEnabled(before, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(disabled, cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Plugins.Enabled) != 0 || !strings.Contains(string(disabled), "note: keep") {
		t.Fatalf("merge override:\n%s\n%#v", disabled, cfg.Plugins.Enabled)
	}
	aliased := []byte("schema_version: 2\ndefaults: &defaults\n  enabled: [fixture]\nplugins:\n  <<: *defaults\n")
	if _, err := PlanPluginEnabled(aliased, nil); err == nil {
		t.Fatal("accepted an aliased plugin merge")
	}
	anchored := []byte("schema_version: 2\nplugins:\n  enabled: &chosen [fixture]\ncustom: *chosen\n")
	if _, err := PlanPluginEnabled(anchored, []string{"fixture", "other"}); err == nil {
		t.Fatal("accepted an anchored enabled list that another key aliases")
	}
	if _, err := PlanPluginEnabled(anchored, nil); err == nil {
		t.Fatal("accepted deletion of an anchored enabled list")
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

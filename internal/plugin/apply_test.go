package plugin

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestApplyWithoutPluginsIsByteIdentical(t *testing.T) {
	original := templates.InstructionEntryFile("AGENTS.md") + "\n## Team\n\nkeep  \n"
	got, err := Apply(original, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != original {
		t.Fatalf("changed bytes\n%s", got)
	}
	plain := "project notes with no contract\n"
	got, err = Apply(plain, nil)
	if err != nil || got != plain {
		t.Fatalf("%q %v", got, err)
	}
}

func TestApplyAddsUpdatesAndRemoves(t *testing.T) {
	original := templates.InstructionEntryFile("CLAUDE.md") + "\n## Team\n\nkeep  \n"
	first := []Block{{ID: "fixture", Body: "Remember the roster.", ConfigPath: ".kit/plugins/fixture/config.yaml"}}
	with, err := Apply(original, first)
	if err != nil {
		t.Fatal(err)
	}
	want := "\n<!-- BEGIN KIT-PLUGIN fixture -->\n<!-- Plugin config: .kit/plugins/fixture/config.yaml -->\nRemember the roster.\n<!-- END KIT-PLUGIN fixture -->\n"
	end := templates.UniversalContractEndMarker + "\n"
	if !strings.Contains(with, end+want+"\n## Team\n\nkeep  \n") {
		t.Fatalf("block:\n%s", with)
	}
	if strings.Count(with, templates.UniversalContractBeginMarker) != 1 {
		t.Fatal("contract count")
	}
	again, err := Apply(with, first)
	if err != nil || again != with {
		t.Fatalf("not idempotent: %v", err)
	}
	second := []Block{{ID: "fixture", Body: "Updated roster."}}
	updated, err := Apply(with, second)
	if err != nil || strings.Contains(updated, "Remember the roster.") || !strings.Contains(updated, "Updated roster.") {
		t.Fatalf("update:\n%s %v", updated, err)
	}
	if strings.Contains(updated, "Plugin config:") {
		t.Fatal("config comment without a config path")
	}
	back, err := Apply(updated, nil)
	if err != nil || back != original {
		t.Fatalf("restore\n%q\n%q\n%v", back, original, err)
	}
}

func TestApplyKeepsEnabledOrder(t *testing.T) {
	original := templates.InstructionEntryFile("AGENTS.md")
	blocks := []Block{{ID: "alpha", Body: "one"}, {ID: "beta", Body: "two"}}
	got, err := Apply(original, blocks)
	if err != nil {
		t.Fatal(err)
	}
	alpha := strings.Index(got, "BEGIN KIT-PLUGIN alpha")
	beta := strings.Index(got, "BEGIN KIT-PLUGIN beta")
	if alpha < 0 || beta < alpha {
		t.Fatalf("order:\n%s", got)
	}
	again, err := Apply(got, blocks)
	if err != nil || again != got {
		t.Fatal("second apply changed")
	}
	back, err := Apply(got, nil)
	if err != nil || back != original {
		t.Fatal("restore failed")
	}
}

func TestApplyRejectsBrokenRegions(t *testing.T) {
	base := templates.InstructionEntryFile("AGENTS.md")
	broken := base + "\n<!-- BEGIN KIT-PLUGIN fixture -->\nno end\n"
	if _, err := Apply(broken, nil); err == nil {
		t.Fatal("accepted an incomplete block")
	}
	if _, err := Apply("no contract\n", []Block{{ID: "fixture", Body: "x"}}); err == nil {
		t.Fatal("accepted a file with no contract")
	}
	if _, err := Apply(base, []Block{{ID: "Bad", Body: "x"}}); err == nil {
		t.Fatal("accepted a bad id")
	}
}

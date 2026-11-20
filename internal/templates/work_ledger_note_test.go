package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestWorkLedgerNoteFollowsConfig(t *testing.T) {
	if WorkLedgerNote(config.WorkLedgerConfig{}) != "" {
		t.Fatal("unconfigured ledger emitted a contract note")
	}
	cfg := config.WorkLedgerConfig{
		Provider:   "plane",
		OwnerLabel: "agent:example",
		Plane:      config.PlaneLedgerConfig{Workspace: "example-workspace", Project: "PROJ", Name: "Ledger"},
	}
	note := WorkLedgerNote(cfg)
	for _, want := range []string{"example-workspace", "Ledger (PROJ)", "agent:example", "repository owner", "does not call the ledger API"} {
		if !strings.Contains(note, want) {
			t.Errorf("note missing %q", want)
		}
	}
	file := ApplyWorkLedgerNote(InstructionEntryFile("AGENTS.md"), cfg)
	if !strings.Contains(file, note) || strings.Count(file, "BEGIN KIT-MANAGED WORK LEDGER") != 1 {
		t.Fatal("note was not inserted once")
	}
	again := ApplyWorkLedgerNote(file, cfg)
	if again != file {
		t.Fatal("applying the note twice changed the file")
	}
	cleared := ApplyWorkLedgerNote(file, config.WorkLedgerConfig{})
	if strings.Contains(cleared, "WORK LEDGER") || cleared != InstructionEntryFile("AGENTS.md") {
		t.Fatal("clearing the provider left a ledger note")
	}
}

func TestGeneratedLedgerOutputHasNoSpecificTarget(t *testing.T) {
	forbidden := []string{"Jameson", "reasoning-machines", "app.plane.so"}
	generated := []string{
		UniversalContractBlock(),
		InstructionEntryFile("AGENTS.md"),
		InstructionEntryFile("CLAUDE.md"),
		InstructionEntryFile(".github/copilot-instructions.md"),
		WorkLedgerNote(config.WorkLedgerConfig{}),
		WorkLedgerNote(config.WorkLedgerConfig{Provider: "plane", Plane: config.PlaneLedgerConfig{Workspace: "<workspace>", Project: "<project>"}}),
	}
	for _, text := range generated {
		for _, bad := range forbidden {
			if strings.Contains(text, bad) {
				t.Errorf("generated output contains %q", bad)
			}
		}
	}
}

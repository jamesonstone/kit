package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

// Ledger writes are native agent actions. This pins the shipped instructions.
func TestEmbeddedWorkLedgerHook(t *testing.T) {
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var delivery, cleanup string
	for _, rule := range registry {
		switch rule.Slug {
		case "delivery":
			delivery = rule.Content
		case "work-cleanup":
			cleanup = rule.Content
		}
	}
	for _, marker := range []string{
		"### Work ledger hook",
		"An omitted `work_ledger` or an omitted `provider` is off",
		"`provider: none` is the same no-op",
		"`plane` is implemented",
		"`linear`, `jira`, and `notion` are reserved",
		"does not call the ledger API",
		"Plane MCP",
		"work_ledger.plane.workspace",
		"work_ledger.plane.project",
		"work_ledger.owner_label",
		"`<workspace>`",
		"`<project>`",
		"The repository owner supplies",
		"only when `work_ledger.provider` is set",
		"configured active state (default `Active`)",
		"configured in-review state (default `In review`)",
		"Do not move that item back to the active state",
		"configured done state (default `Done`)",
		"A failed search is not an empty project",
		"do not by themselves stop the ticket",
	} {
		if !strings.Contains(delivery, marker) {
			t.Errorf("embedded delivery missing %q", marker)
		}
	}
	if !strings.Contains(cleanup, "A configured `work_ledger` follows `delivery` after a satisfying merge and sets Done.") {
		t.Fatal("work-cleanup does not route a configured ledger to Done")
	}
	contract := templates.UniversalContractBlock()
	if strings.Contains(contract, "BEGIN KIT-MANAGED WORK LEDGER") {
		t.Fatal("unconfigured contract emitted a work ledger note")
	}
	generated := []string{delivery, cleanup, contract}
	for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
		data, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		generated = append(generated, string(data))
	}
	for _, bad := range []string{"Jameson", "reasoning-machines", "app.plane.so"} {
		for _, text := range generated {
			if strings.Contains(text, bad) {
				t.Errorf("generated ledger guidance contains %q", bad)
			}
		}
	}
}

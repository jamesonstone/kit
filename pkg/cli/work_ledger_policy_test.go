package cli

import (
	"context"
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
	var delivery string
	for _, rule := range registry {
		if rule.Slug == "delivery" {
			delivery = rule.Content
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
		"workspace `reasoning-machines`, project Work Ledger (`WL`)",
		"set status Active",
		"`agent:<x>`",
		"`grokbot:<bot_name>`",
		"set status In review",
		"set status Done",
		"A failed search is not an empty project",
		"do not by themselves stop the ticket",
	} {
		if !strings.Contains(delivery, marker) {
			t.Errorf("embedded delivery missing %q", marker)
		}
	}
	contract := templates.UniversalContractBlock()
	for _, marker := range []string{"`work_ledger`", "Omitted `work_ledger.provider` is off"} {
		if !strings.Contains(contract, marker) {
			t.Errorf("universal contract missing %q", marker)
		}
	}
}

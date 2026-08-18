package cli

import (
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestReconcileFindsStaleDeletionSafetyGuidance(t *testing.T) {
	tests := []struct {
		name    string
		version int
		path    string
		snippet string
		audit   func(string) []reconcileFinding
	}{
		{
			name:    "V2 RLM route",
			version: config.InstructionScaffoldVersionTOC,
			path:    "docs/agents/RLM.md",
			snippet: "Load `docs/references/rules/deletion-safety.md` before designing deletion behavior or deleting persistent project, user, business, or external-system state",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 root soft-delete default",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "An unqualified delete means soft delete",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root hard-delete confirmation",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "obtain a specific manual confirmation from the human for those exact current targets",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root bounded-selector snapshot",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "materialized target IDs or an immutable snapshot/version token",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root selector materialization order",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "bounded selector first resolved to the exact current target set",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 guardrails route",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/GUARDRAILS.md",
			snippet: "post-outline specific manual confirmation before every hard delete",
			audit:   auditV3SupportGuidance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := writeCurrentReconcileGuidanceFixture(t, tt.version)
			removeGuidanceSnippet(t, projectRoot, tt.path, tt.snippet)
			assertStaleGuidanceFinding(t, projectRoot, tt.path, tt.snippet, tt.audit(projectRoot))
		})
	}
}

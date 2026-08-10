package cli

import (
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestReconcileFindsStaleCrossRepositoryProgramCoordinationGuidance(t *testing.T) {
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
			snippet: "Load `docs/references/rules/cross-repository-program-coordination.md` before implementing or resuming an accepted plan that spans multiple repositories with dependent deliverables, staged deployment or activation, or expected agent or session handoff",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V2 tooling frontier",
			version: config.InstructionScaffoldVersionTOC,
			path:    "docs/agents/TOOLING.md",
			snippet: "When `cross-repository-program-coordination` applies, dispatch only the canonical program ledger's reconciled ready frontier and checkpoint program state after each material transition or handoff",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 root gate",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Before implementing or resuming an accepted plan that spans multiple repositories and includes dependent deliverables, staged deployment or activation, or expected agent or session handoff, load `docs/references/rules/cross-repository-program-coordination.md`.",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 tooling frontier",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/TOOLING.md",
			snippet: "When `cross-repository-program-coordination` applies, dispatch only the canonical program ledger's reconciled ready frontier and checkpoint program state after each material transition or handoff",
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

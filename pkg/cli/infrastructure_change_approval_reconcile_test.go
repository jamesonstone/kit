package cli

import (
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestReconcileFindsStaleInfrastructureChangeApprovalGuidance(t *testing.T) {
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
			snippet: "Load `docs/references/rules/infrastructure-change-approval.md` before planning or performing public-cloud or infrastructure-as-code mutations",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 root gate",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Before mutating public-cloud resources or infrastructure-as-code source, configuration, or state, load `docs/references/rules/infrastructure-change-approval.md`.",
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

package cli

import (
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestReconcileFindsStaleHumanAuthorshipGuidance(t *testing.T) {
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
			snippet: "Load `docs/references/rules/human-authorship.md` before any commit, pull request, issue, review comment, or other attribution text",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 RLM route",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/RLM.md",
			snippet: "Load `docs/references/rules/human-authorship.md` before any commit, pull request, issue, review comment, or other attribution text",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V2 references index",
			version: config.InstructionScaffoldVersionTOC,
			path:    "docs/references/README.md",
			snippet: "`rules/human-authorship.md`",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 references index",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/references/README.md",
			snippet: "`rules/human-authorship.md`",
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

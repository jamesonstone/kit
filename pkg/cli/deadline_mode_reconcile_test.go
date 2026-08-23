package cli

import (
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestReconcileFindsStaleDeadlineModeGuidance(t *testing.T) {
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
			snippet: "Load `docs/references/rules/deadline-mode.md` only when the user explicitly signals a real time constraint or deadline in-thread; never infer or proactively suggest deadline mode",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 RLM route",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/RLM.md",
			snippet: "Load `docs/references/rules/deadline-mode.md` only when the user explicitly signals a real time constraint or deadline in-thread; never infer or proactively suggest deadline mode",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V2 references index",
			version: config.InstructionScaffoldVersionTOC,
			path:    "docs/references/README.md",
			snippet: "`rules/deadline-mode.md`",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 references index",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/references/README.md",
			snippet: "`rules/deadline-mode.md`",
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

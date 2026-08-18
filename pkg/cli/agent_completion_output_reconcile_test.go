package cli

import (
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestReconcileFindsStaleAgentCompletionOutputGuidance(t *testing.T) {
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
			snippet: "Load `docs/references/rules/agent-completion-output.md` before a terminal task completion or handoff response",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V2 references index route",
			version: config.InstructionScaffoldVersionTOC,
			path:    "docs/references/README.md",
			snippet: "`rules/agent-completion-output.md`",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 root status heading",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root action table",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Type | Action required | Why | Continue with",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 guardrails profiles",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/GUARDRAILS.md",
			snippet: "implementation, research, diagnosis, planning, validation, review, operations, coordination, or fallback",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 references index route",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/references/README.md",
			snippet: "`rules/agent-completion-output.md`",
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

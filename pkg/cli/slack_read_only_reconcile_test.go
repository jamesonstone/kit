package cli

import (
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestReconcileFindsStaleSlackReadOnlyGuidance(t *testing.T) {
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
			snippet: "Load `docs/references/rules/slack-read-only.md` before any Slack write, and when a Slack link, channel, thread, or search is part of the task",
			audit:   auditV2SupportGuidance,
		},
		{
			name:    "V3 RLM route",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/RLM.md",
			snippet: "Load `docs/references/rules/slack-read-only.md` before any Slack write, and when a Slack link, channel, thread, or search is part of the task",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root hard gate",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "## Slack: Read-Only by Default, Explicit Approval Required to Send",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 references index",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/references/README.md",
			snippet: "`rules/slack-read-only.md`",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 guardrails hard gate",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/GUARDRAILS.md",
			snippet: "## Slack: Read-Only by Default, Explicit Approval Required to Send",
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

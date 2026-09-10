package cli

import (
	"path/filepath"
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
			snippet: "Load `docs/references/rules/agent-completion-output.md` before a substantial terminal completion or handoff response",
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
			name:    "V3 root free shape",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Write each response, including terminal completions and handoffs, in the shape its content calls for",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root conveyed facts",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "conveys what the user now has, what remains unfinished and why",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root required action",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "the exact command or prompt when there is one",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root plain outcome",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Say plainly whether the work is finished, partly finished, blocked, or failed",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root blocker prominence",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Keep blockers and unfinished scope as prominent as the successes",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root observed check state",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Report each check as observed",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root literal provider states",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE are preserved verbatim",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root ran and passed",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Report a check as passing only when it ran and passed",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root unvalidated disclosure",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "When something could not be validated, say so and say why",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root confidence",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "Distinguish a verified fact from an inference and from a hypothesis",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 root recoverable mutations",
			version: config.InstructionScaffoldVersionMemory,
			path:    "AGENTS.md",
			snippet: "the identifiers a reader needs to find or undo it",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 guardrails literal provider states",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/GUARDRAILS.md",
			snippet: "PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE are preserved verbatim",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 guardrails ran and passed",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/GUARDRAILS.md",
			snippet: "Report a check as passing only when it ran and passed",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 guardrails content not layout",
			version: config.InstructionScaffoldVersionMemory,
			path:    "docs/agents/GUARDRAILS.md",
			snippet: "Satisfy them on content; a heading alone satisfies none of them",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 copilot literal provider states",
			version: config.InstructionScaffoldVersionMemory,
			path:    ".github/copilot-instructions.md",
			snippet: "PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE are preserved verbatim",
			audit:   auditV3SupportGuidance,
		},
		{
			name:    "V3 copilot ran and passed",
			version: config.InstructionScaffoldVersionMemory,
			path:    ".github/copilot-instructions.md",
			snippet: "Report a check as passing only when it ran and passed",
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

func TestAuditV3SupportGuidanceFindsLegacyOperatorActionTable(t *testing.T) {
	projectRoot := writeCurrentReconcileGuidanceFixture(t, config.InstructionScaffoldVersionMemory)
	relativePath := "AGENTS.md"
	absolutePath := filepath.Join(projectRoot, relativePath)
	content := readFile(t, absolutePath)
	writeFile(t, absolutePath, content+"\n"+legacyOperatorActionTableHeader+"\n")
	assertStaleGuidanceFinding(
		t,
		projectRoot,
		relativePath,
		legacyOperatorActionTableHeader,
		auditV3SupportGuidance(projectRoot),
	)
}

func TestAuditV3SupportGuidanceFindsSupersededCompletionEnvelope(t *testing.T) {
	for _, snippet := range []string{legacyStatusHeading, legacyPrioritizedActionList} {
		t.Run(snippet, func(t *testing.T) {
			projectRoot := writeCurrentReconcileGuidanceFixture(t, config.InstructionScaffoldVersionMemory)
			relativePath := "AGENTS.md"
			absolutePath := filepath.Join(projectRoot, relativePath)
			content := readFile(t, absolutePath)
			writeFile(t, absolutePath, content+"\n"+snippet+"\n")
			assertStaleGuidanceFinding(
				t,
				projectRoot,
				relativePath,
				snippet,
				auditV3SupportGuidance(projectRoot),
			)
		})
	}
}

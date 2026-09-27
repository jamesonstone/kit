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

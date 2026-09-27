package cli

import (
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

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

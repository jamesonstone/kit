package cli

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/templates"
)

func TestMergeInstructionFileContent_FailsOnDuplicateRecognizedSections(t *testing.T) {
	existing := `# AGENTS

## Source of truth

first

## Source of truth

second
`

	_, _, err := mergeInstructionFileContent(existing, templates.InstructionFile(agentsMDPath))
	if err == nil || !strings.Contains(err.Error(), "duplicate recognized section") {
		t.Fatalf("expected duplicate recognized section error, got %v", err)
	}
}

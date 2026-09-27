package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionTemplatesRouteDeadlineModeRule(t *testing.T) {
	rlmRoute := "Load `docs/references/rules/deadline-mode.md` only when the user explicitly signals a real time constraint or deadline in-thread; never infer or proactively suggest deadline mode"
	indexRoute := "rules/deadline-mode.md"

	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
	} {
		files := InstructionSupportFiles(version)
		rlm := fileContentByPath(files, "docs/agents/RLM.md")
		if !strings.Contains(rlm, rlmRoute) {
			t.Errorf("scaffold version %d RLM missing %q", version, rlmRoute)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, indexRoute) {
			t.Errorf("scaffold version %d references index missing %q", version, indexRoute)
		}
	}

	if strings.Contains(LegacyConstitutionBaselineSection, "deadline-mode") {
		t.Fatal("Constitution template must not route deadline-mode; it stays conditional and pointer-loaded only")
	}
}

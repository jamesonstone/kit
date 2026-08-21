package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionTemplatesRouteHumanAuthorshipRule(t *testing.T) {
	rlmRoute := "Load `docs/references/rules/human-authorship.md` before any commit, pull request, issue, review comment, or other attribution text"
	indexRoute := "rules/human-authorship.md"
	constitutionRoute := "docs/references/rules/human-authorship.md"

	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
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

	if !strings.Contains(Constitution, constitutionRoute) {
		t.Fatal("Constitution template does not route human-authorship")
	}
}

func TestImplementationDeliverySelectsHumanAuthorshipOptionally(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		if artifact.Slug != "implementation-delivery" {
			continue
		}
		if !strings.Contains(artifact.Content, "slug: human-authorship\n    required: false") {
			t.Fatal("implementation-delivery does not select human-authorship as optional evidence")
		}
		return
	}
	t.Fatal("embedded implementation-delivery workflow not found")
}

package templates

import (
	"strings"
	"testing"
)

func TestBrainstormTemplateIncludesDependenciesTable(t *testing.T) {
	checks := []string{
		"## DEPENDENCIES",
		"| Dependency | Type | Location | Used For | Status |",
		"| none | n/a | n/a | no phase dependencies recorded yet | active |",
	}

	for _, check := range checks {
		if !strings.Contains(BrainstormArtifact, check) {
			t.Fatalf("expected BrainstormArtifact to contain %q", check)
		}
	}
}

func TestInstructionTemplatesRequirePopulatedSections(t *testing.T) {
	checks := []string{
		"## Document Completeness",
		"every required section must be populated",
		"`not applicable`, `not required`, or `no additional information required`",
	}

	templates := map[string]string{
		"AGENTS.md":                       AgentsMD,
		"CLAUDE.md":                       ClaudeMD,
		".github/copilot-instructions.md": CopilotInstructionsMD,
	}

	for name, content := range templates {
		for _, check := range checks {
			if !strings.Contains(content, check) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
	}
}

func TestInstructionTemplatesIncludeReadinessGate(t *testing.T) {
	checks := []string{
		"implementation readiness gate",
		"If the gate fails, update docs first, then code",
	}

	templates := map[string]string{
		"AGENTS.md":                       AgentsMD,
		"CLAUDE.md":                       ClaudeMD,
		".github/copilot-instructions.md": CopilotInstructionsMD,
	}

	for name, content := range templates {
		for _, check := range checks {
			if !strings.Contains(content, check) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
	}
}

func TestSpecTemplateIncludesSkillsAndDependencies(t *testing.T) {
	checks := []string{
		"## SKILLS",
		"| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |",
		"| none | n/a | n/a | no additional skills required | no |",
		"## DEPENDENCIES",
		"| Dependency | Type | Location | Used For | Status |",
		"| none | n/a | n/a | no supporting dependencies recorded yet | active |",
	}

	for _, check := range checks {
		if !strings.Contains(Spec, check) {
			t.Fatalf("expected Spec to contain %q", check)
		}
	}
}

func TestPlanTemplateIncludesDependenciesTable(t *testing.T) {
	checks := []string{
		"## DEPENDENCIES",
		"| Dependency | Type | Location | Used For | Status |",
		"| none | n/a | n/a | no planning dependencies recorded yet | active |",
	}

	for _, check := range checks {
		if !strings.Contains(Plan, check) {
			t.Fatalf("expected Plan to contain %q", check)
		}
	}
}

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

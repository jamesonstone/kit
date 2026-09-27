package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestSpecTemplateUsesV3LivingSpecSections(t *testing.T) {
	checks := []string{
		"## PURPOSE",
		"## CONTEXT",
		"## REQUIREMENTS",
		"## ACCEPTED PLAN",
		"## DECISIONS",
		"## DISCOVERIES",
		"## VALIDATION",
		"## OUTCOME",
		"## REPOSITORY MEMORY",
	}

	for _, check := range checks {
		if !strings.Contains(Spec, check) {
			t.Fatalf("expected Spec to contain %q", check)
		}
	}

	doc := document.Parse(BuildSpecArtifactForFeature(document.FeatureMetadataFromDir("0001-sample")), "SPEC.md", document.TypeSpec)
	if doc.Metadata == nil || doc.Metadata.WorkflowVersion != 3 || doc.Metadata.Phase != "clarify" {
		t.Fatalf("expected generated spec metadata to mark v3 clarify workflow, got %#v", doc.Metadata)
	}
	if _, ok := doc.ClarificationState(); ok {
		t.Fatal("v3 generated spec must not include clarification confidence metadata")
	}
	for _, removed := range []string{"## CLARIFICATIONS", "## TASK CHECKLIST", "## VALIDATION MAP"} {
		if strings.Contains(Spec, removed) {
			t.Fatalf("v3 Spec unexpectedly contains legacy section %q", removed)
		}
	}
}

func TestPlanTemplateUsesReferenceProseSection(t *testing.T) {
	checks := []string{
		"## DEPENDENCIES",
		"References are tracked in front matter.",
	}

	for _, check := range checks {
		if !strings.Contains(Plan, check) {
			t.Fatalf("expected Plan to contain %q", check)
		}
	}
}

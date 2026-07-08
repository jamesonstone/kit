package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func TestInstructionTemplatesIncludeDocAndExportHygiene(t *testing.T) {
	checks := map[string][]string{
		"GUARDRAILS.md": {
			"Always update affected documentation",
			"unused exports",
			"reduce its visibility",
			"attached pasted-text file",
			"self-review and no-known-errors gate",
			"Before staging or committing, self-review the diff",
		},
	}

	for name, snippets := range checks {
		content := fileContentByPath(InstructionSupportFiles(config.InstructionScaffoldVersionTOC), "docs/agents/"+name)
		for _, snippet := range snippets {
			if !strings.Contains(content, snippet) {
				t.Fatalf("expected %s to contain %q", name, snippet)
			}
		}
	}
}

func TestInstructionTemplatesScopeCodeFileSizeGuidance(t *testing.T) {
	legacyChecks := []string{
		"Code file size guideline",
		"implementation/source files around 300 lines",
		"documentation files",
		"`docs/**`",
		"`.kit/**`",
		"`.kit.yaml`",
	}

	for name, content := range map[string]string{
		"AGENTS.md":                       LegacyAgentsMD,
		"CLAUDE.md":                       LegacyClaudeMD,
		".github/copilot-instructions.md": LegacyCopilotInstructionsMD,
	} {
		for _, check := range legacyChecks {
			if !strings.Contains(content, check) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
		for _, stale := range []string{
			"Hard file size limit: 300 lines",
			"Keep files under 300 lines when possible",
		} {
			if strings.Contains(content, stale) {
				t.Fatalf("expected %s not to contain stale unscoped guidance %q", name, stale)
			}
		}
	}

	guardrails := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionTOC),
		"docs/agents/GUARDRAILS.md",
	)
	for _, check := range []string{
		"implementation/source code files around 300 lines",
		"documentation files, `docs/**`, `.kit/**`, or `.kit.yaml`",
	} {
		if !strings.Contains(guardrails, check) {
			t.Fatalf("expected GUARDRAILS.md to contain %q", check)
		}
	}
}

func TestConstitutionTemplateIncludesKitManagedBaselineRules(t *testing.T) {
	for _, check := range []string{
		"### Kit-Managed Baseline Rules",
		"BEGIN KIT-MANAGED BASELINE RULES",
		"docs/notes/<feature>",
		"optional source material, not canonical truth",
		"Prefer implementation/source code files around 300 lines",
		"Do not apply the code-file size guideline to documentation files",
	} {
		if !strings.Contains(Constitution, check) {
			t.Fatalf("expected Constitution template to contain %q", check)
		}
	}
}

func TestReferencesTemplateMentionsFeatureNotesRuleset(t *testing.T) {
	content := fileContentByPath(InstructionSupportFiles(config.InstructionScaffoldVersionTOC), "docs/references/README.md")
	for _, check := range []string{
		"rules/feature-notes.md",
		"docs/notes/<feature>",
		"not canonical truth",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected references README template to contain %q, got:\n%s", check, content)
		}
	}
}

func TestDefaultInstructionTemplatesGlossRLMAndCopilotFallback(t *testing.T) {
	for name, content := range map[string]string{
		"AGENTS.md": AgentsMD,
		"CLAUDE.md": ClaudeMD,
	} {
		if !strings.Contains(content, "just-in-time context loading") {
			t.Fatalf("expected %s to route to RLM guidance on first use", name)
		}
		if !strings.Contains(content, "attached pasted-text file") {
			t.Fatalf("expected %s to include pasted-text attachment guidance", name)
		}
	}

	copilotChecks := []string{
		"Use `docs/agents/RLM.md` when full-context loading would be noisy or wasteful",
		"attached pasted-text file",
		"## Runtime Routing",
		"## Non-Negotiable Rules",
		"Repo-local docs under `docs/` are the source of truth",
		"Do not treat `.claude/skills` as canonical discovery input",
	}

	for _, check := range copilotChecks {
		if !strings.Contains(CopilotInstructionsMD, check) {
			t.Fatalf("expected CopilotInstructionsMD to contain %q", check)
		}
	}
}

func TestDefaultInstructionTemplatesUseTOCModel(t *testing.T) {
	for name, content := range map[string]string{
		"AGENTS.md": AgentsMD,
		"CLAUDE.md": ClaudeMD,
	} {
		for _, check := range []string{
			"routing table",
			"`docs/agents/README.md`",
			"`docs/references/README.md`",
		} {
			if !strings.Contains(strings.ToLower(content), strings.ToLower(check)) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
	}

	for name, content := range map[string]string{
		".github/copilot-instructions.md": CopilotInstructionsMD,
	} {
		for _, check := range []string{
			"`docs/agents/README.md`",
			"`docs/specs/<feature>/`",
		} {
			if !strings.Contains(strings.ToLower(content), strings.ToLower(check)) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
	}
}

func TestSpecTemplateUsesV2SingleArtifactSections(t *testing.T) {
	checks := []string{
		"## THESIS",
		"## CONTEXT",
		"## CLARIFICATIONS",
		"## REQUIREMENTS",
		"## ASSUMPTIONS",
		"## ACCEPTANCE CRITERIA",
		"## IMPLEMENTATION PLAN",
		"## TASK CHECKLIST",
		"## VALIDATION MAP",
		"## REFLECTION NOTES",
		"## DOCUMENTATION UPDATES",
		"## DELIVERY DECISION",
		"## EVIDENCE",
	}

	for _, check := range checks {
		if !strings.Contains(Spec, check) {
			t.Fatalf("expected Spec to contain %q", check)
		}
	}

	doc := document.Parse(BuildSpecArtifactForFeature(document.FeatureMetadataFromDir("0001-sample")), "SPEC.md", document.TypeSpec)
	if doc.Metadata == nil || doc.Metadata.WorkflowVersion != 2 || doc.Metadata.Phase != "clarify" {
		t.Fatalf("expected generated spec metadata to mark v2 clarify workflow, got %#v", doc.Metadata)
	}
	clarification, ok := doc.ClarificationState()
	if !ok || clarification.Status != document.ClarificationStatusOpen {
		t.Fatalf("expected generated spec metadata to include open clarification state, got %#v ok=%v", clarification, ok)
	}
	if confidence, ok := clarification.ConfidenceValue(); !ok || confidence != 0 {
		t.Fatalf("clarification confidence = %d, %v; want 0, true", confidence, ok)
	}
	if unresolved, ok := clarification.UnresolvedQuestionsValue(); !ok || unresolved != 1 {
		t.Fatalf("clarification unresolved = %d, %v; want 1, true", unresolved, ok)
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

func fileContentByPath(files []ScaffoldFile, relativePath string) string {
	for _, file := range files {
		if file.RelativePath == relativePath {
			return file.Content
		}
	}

	return ""
}

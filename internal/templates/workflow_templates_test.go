package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

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
		"docs/references/rules/source-file-size.md",
		"version-control-eligible handwritten implementation/source and test file at 300 physical lines or less",
		"whole-project reconcile and scheduled maintenance audit the entire repository",
		"ignored files, vendored dependencies, and proven generated files",
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
		"version-control-eligible handwritten implementation/source and test file at 300 physical lines or less",
		"whole-project reconcile and scheduled maintenance audit the entire repository",
		"vendored dependencies, and proven generated files",
		"never use minification or arbitrary numbered chunks",
	} {
		if !strings.Contains(Constitution, check) {
			t.Fatalf("expected Constitution template to contain %q", check)
		}
	}
}

func TestReferencesTemplateOmitsRemovedFeatureNotesRuleset(t *testing.T) {
	content := fileContentByPath(InstructionSupportFiles(config.InstructionScaffoldVersionTOC), "docs/references/README.md")
	for _, forbidden := range []string{
		"rules/feature-notes.md",
		"docs/notes/<feature>",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("expected references README template to omit %q, got:\n%s", forbidden, content)
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

func TestInstructionTemplatesIncludeAWSContextHardGate(t *testing.T) {
	for _, version := range []int{config.InstructionScaffoldVersionVerbose, config.InstructionScaffoldVersionTOC} {
		wants := []string{
			"## AWS Context Hard Gate",
			"kit aws verify",
			"before the first AWS-dependent command",
			"before any AWS mutation",
			"verified configured profile explicitly for every AWS-dependent command",
			"including AWS CLI, SDK, Terraform, CDK, deployment, and project scripts",
			"After verification, never use default, another discovered profile, or ambient credentials",
		}
		if version == config.InstructionScaffoldVersionVerbose {
			wants = append(wants,
				"Treat the verified account and ARN as authoritative, not the profile name alone",
				"If verification fails, config is incomplete, credentials are unavailable, or the account mismatches, stop and ask",
			)
		} else {
			wants = append(wants,
				"Treat the verified account and ARN as authoritative; on missing credentials, incomplete config, or mismatch, stop",
			)
		}
		for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
			content := InstructionFileForVersion(path, version)
			for _, want := range wants {
				if !strings.Contains(content, want) {
					t.Fatalf("instruction scaffold version %d %s missing %q", version, path, want)
				}
			}
		}
	}

	guardrails := fileContentByPath(InstructionSupportFiles(config.InstructionScaffoldVersionTOC), "docs/agents/GUARDRAILS.md")
	for _, want := range []string{"## AWS Context Hard Gate", "account ID and ARN as authoritative", "Never fall back to default"} {
		if !strings.Contains(guardrails, want) {
			t.Fatalf("GUARDRAILS.md missing %q", want)
		}
	}
}

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

func TestBuildSpecV2ArtifactRemainsCompatible(t *testing.T) {
	doc := document.Parse(BuildSpecV2ArtifactForFeature(document.FeatureMetadataFromDir("0001-sample")), "SPEC.md", document.TypeSpec)
	if doc.Metadata == nil || doc.Metadata.WorkflowVersion != 2 || doc.Metadata.Phase != "clarify" {
		t.Fatalf("expected V2 compatibility metadata, got %#v", doc.Metadata)
	}
	if !doc.HasSection("THESIS") || !doc.HasSection("VALIDATION MAP") {
		t.Fatalf("expected V2 compatibility sections, got %#v", doc.RequiredSections())
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

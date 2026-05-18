package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func TestFeatureArtifactBuildersIncludeCanonicalFrontMatter(t *testing.T) {
	featureMeta := document.FeatureMetadataFromDir("0001-sample-feature")
	cases := []struct {
		name    string
		docType document.DocumentType
		content string
	}{
		{
			name:    "brainstorm",
			docType: document.TypeBrainstorm,
			content: BuildBrainstormArtifactForFeature("user thesis", featureMeta, []document.MetadataReference{{
				Name:       "Feature notes",
				Type:       "notes",
				Target:     "docs/notes/0001-sample-feature",
				Relation:   document.ReferenceRelationInforms,
				ReadPolicy: document.ReferenceReadPolicyConditional,
				UsedFor:    "optional pre-brainstorm research input",
				Status:     document.ReferenceStatusOptional,
			}}),
		},
		{name: "spec", docType: document.TypeSpec, content: BuildSpecArtifactForFeature(featureMeta)},
		{name: "plan", docType: document.TypePlan, content: BuildPlanArtifactForFeature(featureMeta)},
		{name: "tasks", docType: document.TypeTasks, content: BuildTasksArtifactForFeature(featureMeta)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.Parse(tc.content, tc.name+".md", tc.docType)
			if !doc.FrontMatterPresent {
				t.Fatal("expected generated artifact to include front matter")
			}
			if doc.Metadata == nil || doc.Metadata.Feature.Dir != "0001-sample-feature" || doc.Metadata.Artifact != document.ArtifactForDocumentType(tc.docType) {
				t.Fatalf("unexpected metadata: %#v", doc.Metadata)
			}
			for _, section := range document.RequiredSections[tc.docType] {
				if !doc.HasSection(section) {
					t.Fatalf("expected generated artifact to keep required section %q", section)
				}
			}
		})
	}
}

func TestFeatureArtifactBuildersDoNotDuplicateCanonicalBodyTables(t *testing.T) {
	featureMeta := document.FeatureMetadataFromDir("0001-sample-feature")
	for name, content := range map[string]string{
		"brainstorm": BuildBrainstormArtifactForFeature("user thesis", featureMeta, nil),
		"spec":       BuildSpecArtifactForFeature(featureMeta),
		"plan":       BuildPlanArtifactForFeature(featureMeta),
	} {
		for _, tableHeader := range []string{
			"| Dependency | Type | Location | Used For | Status |",
			"| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |",
		} {
			if strings.Contains(content, tableHeader) {
				t.Fatalf("expected %s builder not to duplicate body table %q", name, tableHeader)
			}
		}
	}
}

func TestBrainstormTemplateUsesReferenceProseSection(t *testing.T) {
	checks := []string{
		"## DEPENDENCIES",
		"References are tracked in front matter.",
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
		"AGENTS.md":                       LegacyAgentsMD,
		"CLAUDE.md":                       LegacyClaudeMD,
		".github/copilot-instructions.md": LegacyCopilotInstructionsMD,
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
		"update the canonical docs first",
	}

	for name, content := range map[string]string{
		"WORKFLOWS.md": fileContentByPath(InstructionSupportFiles(config.InstructionScaffoldVersionTOC), "docs/agents/WORKFLOWS.md"),
	} {
		for _, check := range checks {
			if !strings.Contains(content, check) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
	}
}

func TestInstructionTemplatesDistinguishRLMAndDispatch(t *testing.T) {
	checks := map[string][]string{
		"RLM.md": {
			"RLM is Kit's just-in-time context-routing pattern",
			"Use it for any task where loading full context would be noisy or wasteful",
			"## Runtime Loop",
			"identify the immediate decision",
			"stop loading once the decision is supported",
			"## Context Budget Rules",
			"specific section over full file",
			"docs/PROJECT_PROGRESS_SUMMARY.md",
			"conditional reads only",
			"shared interface or contract",
			"Inspect at most 5 prior feature directories",
			"discovery and context selection first",
			"do not jump straight into parallel execution",
			"Always update affected documentation",
		},
		"TOOLING.md": {
			"Use subagents when the work cleanly separates into low-overlap lanes after discovery",
			"Keep broad or noisy discovery in RLM first",
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

func TestInstructionTemplatesIncludeDocAndExportHygiene(t *testing.T) {
	checks := map[string][]string{
		"GUARDRAILS.md": {
			"Always update affected documentation",
			"unused exports",
			"reduce its visibility",
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

func TestDefaultInstructionTemplatesGlossRLMAndCopilotFallback(t *testing.T) {
	for name, content := range map[string]string{
		"AGENTS.md": AgentsMD,
		"CLAUDE.md": ClaudeMD,
	} {
		if !strings.Contains(content, "just-in-time context loading") {
			t.Fatalf("expected %s to route to RLM guidance on first use", name)
		}
	}

	copilotChecks := []string{
		"Use `docs/agents/RLM.md` when full-context loading would be noisy or wasteful",
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

func TestSpecTemplateIncludesSkillsAndReferenceProse(t *testing.T) {
	checks := []string{
		"## SKILLS",
		"| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |",
		"| none | n/a | n/a | no additional skills required | no |",
		"## DEPENDENCIES",
		"References are tracked in front matter.",
	}

	for _, check := range checks {
		if !strings.Contains(Spec, check) {
			t.Fatalf("expected Spec to contain %q", check)
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

func fileContentByPath(files []ScaffoldFile, relativePath string) string {
	for _, file := range files {
		if file.RelativePath == relativePath {
			return file.Content
		}
	}

	return ""
}

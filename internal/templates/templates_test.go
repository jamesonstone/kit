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
			for _, section := range doc.RequiredSections() {
				if !doc.HasSection(section) {
					t.Fatalf("expected generated artifact to keep required section %q", section)
				}
			}
		})
	}
}

func TestBuildAutoAssignWorkflowRendersSafeGitHubActionsWorkflow(t *testing.T) {
	content := BuildAutoAssignWorkflow([]string{"jamesonstone", "octocat"})

	for _, check := range []string{
		"# Kit-managed auto-assignment workflow.",
		"pull_request_target:",
		"issues: write",
		"pull-requests: read",
		"continue-on-error: true",
		`"jamesonstone"`,
		`"octocat"`,
		"github.rest.issues.addAssignees",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", check, content)
		}
	}
	if strings.Contains(content, "actions/checkout") {
		t.Fatalf("auto-assign workflow must not check out pull request code:\n%s", content)
	}
}

func TestBuildAutoAssignWorkflowNoOpsWithoutAssignees(t *testing.T) {
	content := BuildAutoAssignWorkflow(nil)

	for _, check := range []string{
		"const assignees = [];",
		"No Kit auto-assignees configured; skipping.",
		"continue-on-error: true",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected empty-assignee workflow to contain %q, got:\n%s", check, content)
		}
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
		"every required `SPEC.md` section must be populated",
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
		"v2 readiness gates",
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
			"agent-team-orchestration.md",
			"feature-notes.md",
			"docs/notes/<feature>/README.md",
			"Promote durable conclusions from notes",
			"shared interface or contract",
			"Inspect at most 5 prior feature directories",
			"discovery and context selection first",
			"do not jump straight into parallel execution",
			"Always update affected documentation",
		},
		"TOOLING.md": {
			"## Command Capability Discovery",
			"Use `kit capabilities` when choosing among Kit commands",
			"`docs/references/rules/kit-capabilities-usage.md`",
			"do not maintain Kit's internal command catalog from a downstream project",
			"safe Agent Team Plan",
			"agent-team-orchestration.md",
			"Use subagents when the work cleanly separates into low-overlap lanes after discovery",
			"Default to at most 3 concurrent lanes; never exceed 4",
			"Keep broad or noisy discovery in RLM first",
			"Use `kit pr fix` as the default PR review feedback entrypoint",
			"uses the prompt-producing `kit dispatch --pr` path",
			"does not run the loop agent",
			"post-push reflection cycle before review-thread resolution",
			"resolve matching current unresolved review threads",
			"including human reviewer and CodeRabbit feedback",
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

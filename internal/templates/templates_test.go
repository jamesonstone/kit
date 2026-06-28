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
			"Use subagents when the work cleanly separates into low-overlap lanes after discovery",
			"Keep broad or noisy discovery in RLM first",
			"Use `kit pr fix` as the default PR review repair entrypoint",
			"resolve all matching current unresolved review threads",
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
		"Prefer implementation/source code files around 300 lines",
		"Do not apply the code-file size guideline to documentation files",
	} {
		if !strings.Contains(Constitution, check) {
			t.Fatalf("expected Constitution template to contain %q", check)
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

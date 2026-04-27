package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

func runSpecTemplate(
	specPath, brainstormPath, featureSlug, projectRoot string,
	cfg *config.Config,
	outputOnly bool,
) error {
	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	hasBrainstorm := document.Exists(brainstormPath)
	version := detectInstructionScaffoldVersion(projectRoot, cfg)
	useRLM := specNeedsRLM(featureSlug, specPath, brainstormPath, nil)

	contextRows := [][]string{
		{"CONSTITUTION", fmt.Sprintf("%s — project-wide constraints, principles, priors", constitutionPath)},
	}
	for _, row := range repoInstructionContextRows(projectRoot, cfg) {
		contextRows = append(contextRows, []string{
			row[0],
			fmt.Sprintf("%s — active workflow and skill usage rules", row[1]),
		})
	}
	if hasBrainstorm {
		contextRows = append(contextRows, []string{
			"BRAINSTORM",
			fmt.Sprintf("%s — upstream research context and codebase findings", brainstormPath),
		})
	}
	contextRows = append(contextRows, []string{
		"CANONICAL SKILLS",
		fmt.Sprintf("%s/*/SKILL.md — repo-local reusable skills", cfg.SkillsPath(projectRoot)),
	})
	for _, path := range globalSkillDiscoveryInputs() {
		label := "GLOBAL INPUT"
		if version == config.InstructionScaffoldVersionTOC {
			label = "SECONDARY GLOBAL INPUT"
		}
		contextRows = append(contextRows, []string{
			label,
			fmt.Sprintf("%s — documented global skill or instruction input", path),
		})
	}

	steps := []string{
		"Read CONSTITUTION.md to understand project constraints and principles",
		repoInstructionReadStepText(projectRoot, cfg),
	}
	if hasBrainstorm {
		steps = append(steps, "Read BRAINSTORM.md and carry forward validated findings")
	}
	steps = append(steps,
		"Read the SPEC.md template and understand the required sections",
		relatedFeatureContextStepText(projectRoot, specPath),
		fmt.Sprintf("Analyze only the relevant code and docs surfaced by that filtered prior-work set at %s; widen the search only when the current evidence is insufficient", projectRoot),
	)
	if useRLM {
		steps = append(steps, rlmSpecGuidanceStepText(specPath))
	}
	steps = append(steps,
		"**IMMEDIATELY update SPEC.md** with the context provided above before asking any questions",
		specSkillDiscoveryStepText(projectRoot, cfg, specPath),
		specRelationshipsStepText(specPath),
		specDependencyInventoryStepText(specPath, brainstormPath, hasBrainstorm),
	)
	steps = append(steps, clarificationLoopSteps(
		goalPct,
		fmt.Sprintf(
			"Reassess, save your updates to %s, and continue with additional batches of up to 10 questions until the specification is precise enough to produce a correct, production-quality solution",
			specPath,
		),
	)...)
	steps = append(steps,
		"Continue refining each section of SPEC.md as you learn more:\n"+
			"- PROBLEM: What problem does this feature solve?\n"+
			"- GOALS: What are the measurable outcomes?\n"+
			"- NON-GOALS: What is explicitly out of scope?\n"+
			"- USERS: Who will use this feature?\n"+
			"- SKILLS: Which documented skills should the coding agent use for this feature, where do they live, and when should each one trigger?\n"+
			"- RELATIONSHIPS: Which prior features does this work build on, depend on, or relate to, if any?\n"+
			"- DEPENDENCIES: Which supporting docs, tools, design refs, APIs, libraries, datasets, assets, and other inputs shaped this specification?\n"+
			"- REQUIREMENTS: What must be true for this feature to be complete?\n"+
			"- ACCEPTANCE: How do we verify the feature works?\n"+
			"- EDGE-CASES: What unusual scenarios must be handled?",
	)

	prompt := renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Please review and complete the specification at %s.", specPath))
		doc.Paragraph(fmt.Sprintf("This is a new feature: %s", featureSlug))
		doc.Heading(2, "Context Docs (read first)")
		doc.Table([]string{"File", "Purpose"}, contextRows)
		doc.Heading(2, "Context Provided by User")
		doc.Raw(`<!-- ⚠️ FILL THIS OUT BEFORE SUBMITTING TO YOUR CODING AGENT -->

**PROBLEM**:
<!-- What problem does this feature solve? -->

**GOALS**:
<!-- What are the measurable outcomes? (comma-separated) -->

**NON-GOALS**:
<!-- What is explicitly out of scope? -->

**USERS**:
<!-- Who will use this feature? -->

**REQUIREMENTS**:
<!-- What must be true for this feature to be complete? -->

**ACCEPTANCE**:
<!-- How do we verify the feature works? -->

**EDGE-CASES**:
<!-- What unusual scenarios must be handled? -->`)
		if useRLM {
			doc.Heading(1, "Use RLM Pattern")
		}
		doc.Heading(2, "Your Task")
		doc.OrderedList(1, steps...)
		doc.Paragraph(fmt.Sprintf(
			"Do NOT treat SPEC.md as complete until confidence reaches ≥%d%% and unresolved assumptions = 0.",
			goalPct,
		))
		doc.Heading(2, "SUMMARY Section (MANDATORY)")
		doc.Paragraph(fmt.Sprintf("Once you reach ≥%d%% confidence, write a SUMMARY section at the top of SPEC.md:", goalPct))
		doc.BulletList(
			"1-2 sentences maximum",
			"Information-dense: include the core problem, solution approach, and key constraint",
			"Written for a coding agent who needs to quickly understand the feature",
			`Example: "Adds CSV export for user data with streaming support for large datasets (>100k rows). Must complete in <5s and handle Unicode."`,
		)
		doc.Heading(2, "Rules")
		doc.BulletList(
			"Keep language precise",
			"Avoid implementation details (focus on WHAT, not HOW)",
			"the ## SKILLS section is mandatory and must be populated before sign-off",
			"the ## RELATIONSHIPS section is mandatory and must be set to none or explicit bullets using canonical feature identifiers before sign-off",
			"the ## DEPENDENCIES section must be current before sign-off and must keep exact locations for external design inputs",
			"use repo-local docs and canonical skills first during the skills discovery phase",
			"treat documented global inputs as secondary context after repo-local docs are exhausted",
			"keep the selected skill set minimal and actionable",
			"do not use .claude/skills as canonical discovery input",
			"Spec gate: unresolved assumptions = 0 before sign-off; if unresolved assumptions remain, stop and resolve before marking SPEC complete",
			"Ensure the spec respects constraints defined in CONSTITUTION.md",
			"PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times",
		)
		doc.Raw(renderNonEmptySectionRules("`SPEC.md`"))
	})

	if !outputOnly {
		fmt.Println()
		fmt.Println(dim + "⚠️ IMPORTANT: Before submitting this prompt, fill in the context section" + reset)
		fmt.Println(dim + "   with details about your feature. The more context you provide, the" + reset)
		fmt.Println(dim + "   better the agent can help you write the specification." + reset)
		fmt.Println()
		fmt.Println(dim + "   Tip: Run 'kit spec <feature> --interactive' for a guided" + reset)
		fmt.Println(dim + "   editor-first experience, or add '--inline' for terminal multiline entry." + reset)
		fmt.Println()
	}

	if err := outputPromptForFeatureWithClipboardDefault(prompt, filepath.Dir(specPath), outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		printNumberedNextSteps([]string{
			fmt.Sprintf("Edit %s to define the specification", specPath),
			fmt.Sprintf("Run 'kit plan %s' to create the implementation plan", featureSlug),
		})
	}

	return nil
}

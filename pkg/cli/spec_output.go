package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

func outputCompiledPrompt(
	specPath, brainstormPath, featureSlug, projectRoot string,
	cfg *config.Config,
	answers *specAnswers,
	outputOnly bool,
) error {
	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	hasBrainstorm := document.Exists(brainstormPath)
	version := detectInstructionScaffoldVersion(projectRoot, cfg)

	contextItems := []string{}
	if answers.Problem != "" {
		contextItems = append(contextItems, fmt.Sprintf("**PROBLEM**: %s", answers.Problem))
	}
	if answers.Goals != "" {
		contextItems = append(contextItems, fmt.Sprintf("**GOALS**: %s", answers.Goals))
	}
	if answers.NonGoals != "" {
		contextItems = append(contextItems, fmt.Sprintf("**NON-GOALS**: %s", answers.NonGoals))
	}
	if answers.Users != "" {
		contextItems = append(contextItems, fmt.Sprintf("**USERS**: %s", answers.Users))
	}
	if answers.Requirements != "" {
		contextItems = append(contextItems, fmt.Sprintf("**REQUIREMENTS**: %s", answers.Requirements))
	}
	if answers.Acceptance != "" {
		contextItems = append(contextItems, fmt.Sprintf("**ACCEPTANCE**: %s", answers.Acceptance))
	}
	if answers.EdgeCases != "" {
		contextItems = append(contextItems, fmt.Sprintf("**EDGE-CASES**: %s", answers.EdgeCases))
	}

	hasContext := answers.Problem != "" || answers.Goals != "" || answers.NonGoals != "" ||
		answers.Users != "" || answers.Requirements != "" || answers.Acceptance != "" ||
		answers.EdgeCases != ""
	useRLM := specNeedsRLM(featureSlug, specPath, brainstormPath, answers)

	contextRows := [][]string{{"CONSTITUTION", constitutionPath}}
	contextRows = append(contextRows, repoInstructionContextRows(projectRoot, cfg)...)
	if hasBrainstorm {
		contextRows = append(contextRows, []string{"BRAINSTORM", brainstormPath})
	}
	contextRows = append(contextRows, []string{"SPEC", specPath})
	contextRows = append(contextRows, specSkillDiscoveryContextRows(projectRoot, cfg)...)
	contextRows = append(contextRows, []string{"Project Root", projectRoot})

	steps := []string{
		fmt.Sprintf("Read CONSTITUTION.md (file: %s) to understand project constraints and principles", constitutionPath),
		repoInstructionReadStepText(projectRoot, cfg),
	}
	if hasBrainstorm {
		steps = append(steps, fmt.Sprintf(
			"Read BRAINSTORM.md (file: %s) and treat it as upstream research context",
			brainstormPath,
		))
	}
	steps = append(steps,
		fmt.Sprintf("Read the current SPEC.md (file: %s) and understand the required sections", specPath),
		fmt.Sprintf("Analyze the codebase at %s to understand existing patterns", projectRoot),
	)
	if useRLM {
		steps = append(steps, rlmSpecGuidanceStepText(specPath))
	}
	if hasContext {
		steps = append(steps, fmt.Sprintf(
			"**IMMEDIATELY write all context above into the SPEC.md file at %s** — do NOT ask questions before doing this",
			specPath,
		))
	}
	steps = append(steps,
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
	finalWriteStep := fmt.Sprintf(
		"**Write your findings directly to %s** as you fill in each section:\n"+
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
		specPath,
	)
	if hasContext {
		finalWriteStep = "Continue refining each section of SPEC.md as you learn more:\n" +
			strings.TrimPrefix(finalWriteStep, fmt.Sprintf("**Write your findings directly to %s** as you fill in each section:\n", specPath))
	}
	if hasBrainstorm {
		finalWriteStep += "\n- Carry forward validated findings from BRAINSTORM.md into SPEC.md"
	}
	steps = append(steps, finalWriteStep)

	prompt := renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf(
			"You MUST update the specification file at:\n**File**: %s\n**Feature**: %s\n**Project Root**: %s",
			specPath,
			featureSlug,
			projectRoot,
		))
		doc.Heading(2, "Context Provided by User")
		if len(contextItems) > 0 {
			doc.Raw(strings.Join(contextItems, "\n\n"))
		}
		doc.Heading(2, "Context Docs (read first)")
		doc.Table([]string{"File", "Purpose"}, contextRows)
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
		doc.Heading(2, "IMPORTANT: File Update Requirement")
		doc.Paragraph(fmt.Sprintf(
			"All specification content MUST be written to: %s\nThis file is the single source of truth for this feature. Do not leave content only in chat — persist everything to the file.",
			specPath,
		))
		doc.Heading(2, "Rules")
		rules := []string{
			"Keep language precise",
			"Avoid implementation details (focus on WHAT, not HOW)",
			"the ## SKILLS section is mandatory and must be populated before sign-off",
			"the ## RELATIONSHIPS section is mandatory and must be set to none or explicit bullets using canonical feature identifiers before sign-off",
			"the ## DEPENDENCIES section must be current before sign-off and must keep exact locations for external design inputs",
			"use repo-local docs and canonical skills first during the skills discovery phase",
			"keep the selected skill set minimal and actionable",
			"do not use .claude/skills as canonical discovery input",
			"Spec gate: unresolved assumptions = 0 before sign-off; if unresolved assumptions remain, stop and resolve before marking SPEC complete",
			"Ensure the spec respects constraints defined in CONSTITUTION.md",
			"PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times",
		}
		if version == config.InstructionScaffoldVersionTOC {
			rules = append(rules,
				"route through `docs/agents/README.md` and only read the linked docs relevant to this feature",
				"treat documented global inputs as secondary context after repo-local docs are exhausted",
			)
		}
		doc.BulletList(rules...)
		doc.Raw(renderNonEmptySectionRules("`SPEC.md`"))
	})

	if err := outputPromptWithClipboardDefault(prompt, outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		printNumberedNextSteps([]string{
			"Paste the copied prompt into your coding agent",
			"Work with the agent to refine the specification",
			fmt.Sprintf("Run 'kit plan %s' to create the implementation plan", featureSlug),
		})
	}

	return nil
}

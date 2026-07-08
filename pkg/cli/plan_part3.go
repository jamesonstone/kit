package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

func buildStandardPlanPrompt(
	planPath string,
	specPath string,
	brainstormPath string,
	feat *feature.Feature,
	cfg *config.Config,
	projectRoot string,
) string {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	goalPct := cfg.GoalPercentage
	hasBrainstorm := document.Exists(brainstormPath)

	steps := []string{
		fmt.Sprintf("Read CONSTITUTION.md (file: %s) to understand project constraints and principles", constitutionPath),
	}
	if hasBrainstorm {
		steps = append(steps, fmt.Sprintf("Read BRAINSTORM.md (file: %s) for upstream research context", brainstormPath))
	}
	steps = append(steps,
		fmt.Sprintf("Read SPEC.md (file: %s) fully and treat it as a fixed contract", specPath),
		relatedFeatureContextStepText(projectRoot, planPath),
		fmt.Sprintf("Review the PLAN.md (file: %s) template and required sections", planPath),
		"Identify any missing design decisions required for execution",
		planDependencyInventoryStepText(planPath, specPath, brainstormPath, hasBrainstorm),
	)
	steps = append(steps, clarificationLoopSteps(
		goalPct,
		"Reassess and continue with additional batches of up to 10 questions until the plan is precise enough to produce a correct, production-quality implementation",
	)...)
	steps = append(steps, "Commit to concrete design decisions that make execution unambiguous")

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph("Please review and complete the implementation plan.")
		doc.Heading(2, "File References")
		rows := [][]string{{"CONSTITUTION", constitutionPath}}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM", brainstormPath})
		}
		rows = append(rows,
			[]string{"SPEC", specPath},
			[]string{"PLAN", planPath},
			[]string{"Feature", feat.Slug},
			[]string{"Project Root", projectRoot},
		)
		doc.Table([]string{"Document", "Path"}, rows)
		doc.Paragraph("Your task:")
		doc.OrderedList(1, steps...)
		doc.Paragraph(fmt.Sprintf(
			"Before you write or update PLAN.md:\n- after each batch of up to 10 questions, output your current percentage understanding of the implementation plan so the user can see progress\n- do NOT treat PLAN.md as complete until confidence reaches ≥%d%%",
			goalPct,
		))
		doc.Paragraph("For each section, write only what is required to enable clear task breakdown:")
		doc.Raw(`- SUMMARY
  - one-paragraph overview of the chosen approach

- APPROACH
  - high-level strategy
  - explicit tradeoff decisions
  - no code

- COMPONENTS
  - logical components/modules
  - clear responsibility boundaries

- DATA
  - data shapes, enums, tables, files
  - no schema or serialization code unless unavoidable

- INTERFACES
  - commands, inputs, outputs, side effects
  - files and artifacts touched

- DEPENDENCIES
  - the docs, tools, design refs, APIs, libraries, datasets, assets, and other resources shaping the implementation strategy
  - keep exact URLs or file/node refs in front matter references
  - use status = active, optional, or stale

- RISKS
  - top technical or design risks
  - mitigation per risk

- TESTING
  - validation strategy
  - test types, not test code`)
		doc.Heading(2, "Rules")
		doc.BulletList(
			docsOnlyWorkflowRule("PLAN.md and supporting documentation"),
			"focus on HOW, not WHAT",
			"use BRAINSTORM.md as research context only; SPEC.md remains the binding contract",
			"do not restate requirements",
			"do not introduce new scope beyond SPEC.md",
			"canonical front matter `references` must be current before sign-off and must keep exact targets and stable selectors for external design inputs",
			"do not write tasks",
			"keep language dense and factual",
			"Plan gate: acceptance criteria must be testable and mapped to explicit evidence in PLAN.md before sign-off",
			"ensure plan respects constraints defined in CONSTITUTION.md",
			"PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times",
		)
		doc.Paragraph("The output of PLAN.md must make TASKS.md obvious and deterministic.")
		doc.Raw(renderNonEmptySectionRules("`PLAN.md`"))
		addFinalResponseContract(doc, planFinalResponseContract(feat.Slug)...)
	})
}

// outputWarpPlanPrompt outputs a prompt for Warp coding agent to fill PLAN.md from Warp plan.
func outputWarpPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	goalPct := cfg.GoalPercentage
	hasBrainstorm := document.Exists(brainstormPath)

	steps := []string{
		fmt.Sprintf("Read CONSTITUTION.md (file: %s) to understand project constraints and principles", constitutionPath),
	}
	if hasBrainstorm {
		steps = append(steps, fmt.Sprintf("Read BRAINSTORM.md (file: %s) for upstream research context", brainstormPath))
	}
	steps = append(steps,
		"Read the Warp plan you created and extract the key design decisions",
		fmt.Sprintf("Read SPEC.md (file: %s) to ensure alignment with requirements", specPath),
		relatedFeatureContextStepText(projectRoot, planPath),
		planDependencyInventoryStepText(planPath, specPath, brainstormPath, hasBrainstorm),
		fmt.Sprintf(
			"Fill out each section of PLAN.md (file: %s), adding implementation details beyond what's in the Warp plan:\n"+
				"- SUMMARY: one-paragraph overview (expand from Warp plan's high-level description)\n"+
				"- APPROACH: detailed strategy and tradeoff decisions\n"+
				"- COMPONENTS: logical modules with clear responsibility boundaries\n"+
				"- DATA: data shapes, structures, and storage decisions\n"+
				"- INTERFACES: commands, inputs, outputs, side effects\n"+
				"- DEPENDENCIES: prose summary of the resources that shape the implementation strategy; canonical pointers belong in front matter `references`\n"+
				"- RISKS: technical risks with mitigation strategies\n"+
				"- TESTING: validation strategy and test types",
			planPath,
		),
		"Ensure PLAN.md has MORE detail than the Warp plan — it should make task breakdown obvious",
	)
	steps = append(steps, clarificationLoopSteps(
		goalPct,
		"Reassess and continue with additional batches of up to 10 questions until PLAN.md is precise enough to produce a correct, production-quality implementation",
	)...)

	prompt := renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("I have created a Warp plan for the feature: %s", feat.Slug))
		doc.Heading(2, "File References")
		rows := [][]string{{"CONSTITUTION", constitutionPath}}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM", brainstormPath})
		}
		rows = append(rows,
			[]string{"SPEC", specPath},
			[]string{"PLAN", planPath},
			[]string{"Project Root", projectRoot},
		)
		doc.Table([]string{"Document", "Path"}, rows)
		doc.Paragraph(fmt.Sprintf(
			"Please take the Warp plan you just generated and use it to fill out the PLAN.md document at:\n%s",
			planPath,
		))
		doc.Paragraph("Your task:")
		doc.OrderedList(1, steps...)
		doc.Paragraph(fmt.Sprintf(
			"After completing PLAN.md:\n- state your confidence level that TASKS.md can be derived unambiguously\n- do NOT treat PLAN.md as complete until confidence reaches ≥%d%%",
			goalPct,
		))
		doc.Heading(2, "Rules")
		doc.BulletList(
			docsOnlyWorkflowRule("PLAN.md and supporting documentation"),
			"focus on HOW, not WHAT (SPEC covers WHAT)",
			"use BRAINSTORM.md as research context only; SPEC.md remains the binding contract",
			"do not restate requirements verbatim",
			"do not introduce new scope beyond the Warp plan and SPEC.md",
			"canonical front matter `references` must be current before sign-off and must keep exact targets and stable selectors for external design inputs",
			"keep language dense and factual",
			"Plan gate: acceptance criteria must be testable and mapped to explicit evidence in PLAN.md before sign-off",
			"ensure plan respects constraints defined in CONSTITUTION.md",
			"PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature",
		)
		doc.Paragraph("The output of PLAN.md must make TASKS.md obvious and deterministic.")
		doc.Raw(renderNonEmptySectionRules("`PLAN.md`"))
		addFinalResponseContract(doc, planFinalResponseContract(feat.Slug)...)
	})

	if !outputOnly {
		fmt.Println()
		fmt.Println(whiteBold + "Warp Plan Integration" + reset)
		fmt.Println(dim + "The following files have been created:" + reset)
		fmt.Printf("  • PLAN.md: %s\n", planPath)
		fmt.Printf("  • SPEC.md: %s\n\n", specPath)
	}

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, planCopy); err != nil {
		return fmt.Errorf("failed to output prompt: %w", err)
	}

	return nil
}

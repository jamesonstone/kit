package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

func selectFeatureForPlanPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "SPEC.md")) &&
			document.Exists(filepath.Join(f.Path, "PLAN.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no plans available to regenerate prompts for\n\nRun 'kit legacy plan <feature>' first")
	}

	printSelectionHeader("Select a feature to regenerate the plan prompt for:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(candidates) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := candidates[num-1]
	return &selected, nil
}

// outputStandardPlanPrompt outputs the standard coding agent prompt.
func outputStandardPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	prompt := buildStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, projectRoot)

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, planCopy); err != nil {
		return fmt.Errorf("failed to output prompt: %w", err)
	}

	return nil
}
func buildStandardPlanPrompt(
	planPath string,
	specPath string,
	brainstormPath string,
	feat *feature.Feature,
	cfg *config.Config,
	projectRoot string,
) string {
	constitutionPath := filepath.Join(projectRoot, cfg.ConstitutionPath)
	hasBrainstorm := document.Exists(brainstormPath)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Complete the legacy implementation plan for feature `%s`. This is documentation-only; do not implement product code.", feat.Slug))
		doc.Heading(2, "Context")
		rows := [][]string{
			{"SPEC.md", specPath, "Binding requirements and acceptance"},
			{"PLAN.md", planPath, "Artifact to update"},
			{"Constitution", constitutionPath, "Project invariants"},
			{"Project root", projectRoot, "Discover existing implementation patterns"},
		}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM.md", brainstormPath, "Non-binding research context"})
		}
		doc.Table([]string{"Input", "Path", "Use"}, rows)

		doc.Heading(2, "Planning Contract")
		doc.OrderedList(1,
			"Read SPEC.md as fixed scope. Inspect the smallest relevant code, tests, docs, and prior-feature context needed to ground design decisions; do not invent files or APIs.",
			"Resolve repository-discoverable gaps yourself. Ask concise numbered questions only for a material non-discoverable design choice, with a recommended default and impact; stop before writing a final plan while such a choice remains.",
			fmt.Sprintf("Update `%s` directly with the simplest viable approach, explicit tradeoffs, components/responsibilities, data and interfaces, exact dependencies/references, touched surfaces, sequencing, risks/rollback, and validation strategy.", planPath),
			"Map every binding acceptance criterion to implementation responsibility and evidence. Keep exact external and repo references in front matter.",
			"Make the resulting task breakdown deterministic without writing TASKS.md or implementation code.",
		)

		doc.Heading(2, "Success Criteria")
		doc.BulletList(
			fmt.Sprintf("Confidence is at least %d and no material design question remains.", cfg.GoalPercentage),
			"PLAN.md adds implementation strategy rather than restating requirements, introduces no new scope, and follows repository invariants.",
			"Each planned surface, risk, test, and documentation obligation traces to SPEC.md acceptance and has a concrete evidence method.",
			"Empty optional sections state `not applicable`; placeholder comments are removed and the project progress summary remains accurate.",
		)

		doc.Heading(2, "Output")
		doc.BulletList(
			"Update PLAN.md and supporting documentation only.",
			"Report key decisions, validation strategy, remaining risk, and the next legacy task-generation step.",
		)
		addFinalResponseContract(doc, planFinalResponseContract(feat.Slug)...)
	})
}

func outputWarpPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	prompt := buildWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, projectRoot)

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

func buildWarpPlanPrompt(
	planPath string,
	specPath string,
	brainstormPath string,
	feat *feature.Feature,
	cfg *config.Config,
	projectRoot string,
) string {
	constitutionPath := filepath.Join(projectRoot, cfg.ConstitutionPath)
	hasBrainstorm := document.Exists(brainstormPath)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Convert the Warp plan in the current conversation into the legacy PLAN.md for feature `%s`. This is documentation-only; do not implement product code.", feat.Slug))
		doc.Heading(2, "Context")
		rows := [][]string{
			{"Warp plan", "Current conversation", "Non-binding design input"},
			{"SPEC.md", specPath, "Binding requirements and acceptance"},
			{"PLAN.md", planPath, "Artifact to update"},
			{"Constitution", constitutionPath, "Project invariants"},
			{"Project root", projectRoot, "Discover implementation patterns"},
		}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM.md", brainstormPath, "Non-binding research context"})
		}
		doc.Table([]string{"Input", "Path", "Use"}, rows)

		doc.Heading(2, "Planning Contract")
		doc.OrderedList(1,
			"Extract the Warp plan's concrete design decisions, then verify them against SPEC.md, the constitution, and the smallest relevant code and test surfaces. SPEC.md wins on conflict.",
			"Resolve repository-discoverable gaps yourself. Ask concise numbered questions only for a material non-discoverable design choice, with a recommended default and impact; stop before finalizing while such a choice remains.",
			fmt.Sprintf("Update `%s` directly with the simplest viable approach, explicit tradeoffs, components/responsibilities, data and interfaces, exact dependencies/references, touched surfaces, sequencing, risks/rollback, and validation strategy.", planPath),
			"Map every binding acceptance criterion to implementation responsibility and evidence. Keep exact external and repository references in front matter.",
			"Add implementation detail where the Warp plan is abstract, but introduce no scope beyond SPEC.md and do not write TASKS.md or product code.",
		)

		doc.Heading(2, "Success Criteria")
		doc.BulletList(
			fmt.Sprintf("Confidence is at least %d and no material design question remains.", cfg.GoalPercentage),
			"PLAN.md adds implementation strategy beyond the Warp plan without restating requirements or changing binding scope.",
			"Each planned surface, risk, test, and documentation obligation traces to SPEC.md acceptance and has concrete evidence.",
			"Empty optional sections state `not applicable`; placeholder comments are removed and the project progress summary remains accurate.",
		)

		doc.Heading(2, "Output")
		doc.BulletList(
			"Update PLAN.md and supporting documentation only.",
			"Report key decisions carried forward or changed, validation strategy, remaining risk, and the next legacy task-generation step.",
		)
		addFinalResponseContract(doc, planFinalResponseContract(feat.Slug)...)
	})
}

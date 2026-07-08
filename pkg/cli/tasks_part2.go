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

func buildTasksPrompt(feat *feature.Feature, projectRoot string, cfg *config.Config) string {

	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	hasBrainstorm := document.Exists(brainstormPath)
	goalPct := cfg.GoalPercentage
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	steps := []string{
		fmt.Sprintf("Read CONSTITUTION.md (file: %s) to understand project constraints and principles", constitutionPath),
	}
	if hasBrainstorm {
		steps = append(steps, fmt.Sprintf("Read BRAINSTORM.md (file: %s) to preserve upstream research context", brainstormPath))
	}
	steps = append(steps,
		fmt.Sprintf("Read SPEC.md (file: %s) and PLAN.md (file: %s) fully and treat them as fixed inputs", specPath, planPath),
		fmt.Sprintf("Review the TASKS.md (file: %s) template and required sections", tasksPath),
		fmt.Sprintf("Update TASKS.md directly at %s; do not leave the task breakdown only in chat", tasksPath),
		"Derive an atomic, ordered task list that can be executed without ambiguity",
		"Identify missing decisions that block task generation",
	)
	steps = append(steps, clarificationLoopSteps(
		goalPct,
		"Reassess and continue with additional batches of up to 10 questions until the task plan is precise enough to produce a correct, production-quality implementation",
	)...)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph("Please review and complete the task plan.")
		doc.Heading(2, "File References")
		rows := [][]string{{"CONSTITUTION", constitutionPath}}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM", brainstormPath})
		}
		rows = append(rows,
			[]string{"SPEC", specPath},
			[]string{"PLAN", planPath},
			[]string{"TASKS", tasksPath},
			[]string{"Feature", feat.Slug},
			[]string{"Project Root", projectRoot},
		)
		doc.Table([]string{"Document", "Path"}, rows)
		doc.Paragraph("Your task:")
		doc.OrderedList(1, steps...)
		doc.Paragraph(fmt.Sprintf(
			"Before you write or update TASKS.md:\n- after each batch of up to 10 questions, output your current percentage understanding of the task plan so the user can see progress\n- do NOT treat TASKS.md as complete until confidence reaches ≥%d%%",
			goalPct,
		))
		doc.Heading(2, "TASKS.md Requirements")
		doc.Raw(`A) PROGRESS TABLE (ALWAYS FIRST)
- Fill the top table with one row per task
- Use stable IDs (T001, T002, …)
- STATUS ∈ todo | doing | blocked | done
- OWNER is always "agent"
- DEPENDENCIES lists task IDs (comma-separated) or empty

Table columns:
| ID | TASK | STATUS | OWNER | DEPENDENCIES |

B) TASK LIST (MANDATORY - uses markdown checkboxes)
- Use markdown checkboxes for tracking: - [ ] incomplete, - [x] complete
- Format: - [ ] T001: task description
- This enables 'kit status' to parse progress automatically

C) TASK DETAILS SECTION
For each task ID, include a short block:

### T00X
- GOAL: one sentence outcome
- SCOPE: tight bullets, no fluff
- ACCEPTANCE: concrete checks (what must be true)
- NOTES: only if necessary

D) DEPENDENCIES SECTION
- list any cross-task or external blockers
- include the exact missing decision if applicable
- if there are no blockers or ordering notes, replace placeholder comments with "no additional information required" or "not applicable"

E) NOTES SECTION
- only if required to prevent ambiguity
- otherwise write "not required"

F) PLAN LINKS (OPTIONAL)
- Link to PLAN sections using anchors from headings (lowercase, dashes)
- Examples: [PLAN-APPROACH], [PLAN-COMPONENTS], [PLAN-DATA], [PLAN-RISKS]
- Include in task descriptions to trace back to plan requirements`)
		doc.Heading(2, "Rules")
		doc.BulletList(
			docsOnlyWorkflowRule("TASKS.md and supporting documentation"),
			"focus on executable steps, not prose",
			"use BRAINSTORM.md as research context only; SPEC.md and PLAN.md remain the binding inputs",
			"do not invent new requirements or scope beyond SPEC.md",
			"tasks must map back to PLAN items via section anchors",
			"tasks must imply an unambiguous implementation order",
			"Tasks gate: each task must include an explicit done-condition and required evidence artifact(s) before sign-off",
			"keep language dense and factual",
			"ensure tasks respect constraints defined in CONSTITUTION.md",
			"PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times",
		)
		doc.Paragraph("Output goal:\n- a task list that a coding agent can execute linearly with minimal back-and-forth")
		doc.Raw(renderNonEmptySectionRules("`TASKS.md`"))
		addFinalResponseContract(doc, tasksFinalResponseContract(feat.Slug)...)
	})
}

// selectFeatureForTasks shows an interactive numbered list of features
// that have SPEC.md and PLAN.md but no TASKS.md yet.
func selectFeatureForTasks(specsDir string) (*feature.Feature, error) {
	candidates, err := workflowStageCandidates(specsDir, workflowSelectionStageTasks)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no legacy staged features ready for tasks (need SPEC.md + PLAN.md without TASKS.md)\n\nRun 'kit legacy plan <feature>' to create a plan first")
	}

	printSelectionHeader("Select a feature to create tasks for:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s\n", i+1, f.DirName)
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

func selectFeatureForTasksPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "SPEC.md")) &&
			document.Exists(filepath.Join(f.Path, "PLAN.md")) &&
			document.Exists(filepath.Join(f.Path, "TASKS.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no task plans available to regenerate prompts for\n\nRun 'kit legacy tasks <feature>' first")
	}

	printSelectionHeader("Select a feature to regenerate the tasks prompt for:")
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

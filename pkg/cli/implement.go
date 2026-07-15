package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

var implementCopy bool
var implementOutputOnly bool

var implementCmd = &cobra.Command{
	Use:   "implement [feature]",
	Short: "Deprecated v1 staged workflow: output implementation context",
	Long: `Deprecated v1 staged workflow: run the implementation readiness gate and output a comprehensive
summary for coding agents to begin implementation.

The default workflow implements from an accepted native plan and keeps material
repository memory current. Use this command only for the legacy staged artifact
flow.

Provides:
  - Implementation-readiness gate instructions
  - Feature overview and current status
  - Document reference table (SPEC, PLAN, TASKS)
  - Clear instructions for executing tasks

If no feature is specified, shows an interactive selection of legacy
implement-phase features with incomplete TASKS.md checkboxes.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runImplement,
}

func init() {
	implementCmd.Flags().BoolVar(&implementCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	implementCmd.Flags().BoolVar(&implementOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(implementCmd)
	legacyCmd.AddCommand(implementCmd)
}

func runImplement(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	var feat *feature.Feature

	if len(args) == 0 {
		// interactive mode: select from features ready for implementation
		feat, err = selectFeatureForImplementation(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {
		// direct mode: resolve feature by name
		featureRef := args[0]
		feat, err = loadFeatureWithState(specsDir, cfg, featureRef)
		if err != nil {
			return fmt.Errorf("feature '%s' not found", featureRef)
		}
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")

	// verify all documents exist
	if !document.Exists(specPath) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}
	if !document.Exists(planPath) {
		return fmt.Errorf("PLAN.md not found. Run 'kit legacy plan %s' first", feat.Slug)
	}
	if !document.Exists(tasksPath) {
		return fmt.Errorf("TASKS.md not found. Run 'kit legacy tasks %s' first", feat.Slug)
	}

	wasPaused := feat.Paused
	if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
		return err
	}
	if err := updateRollupForResume(projectRoot, cfg, feat.DirName, wasPaused); err != nil {
		return err
	}
	if wasPaused && !outputOnly {
		fmt.Println("  ✓ Cleared paused state")
	}

	// extract summary from spec
	summary, _ := feature.ExtractSpecSummary(specPath)

	// get task progress
	progress, _ := feature.ParseTaskProgress(tasksPath)

	return outputImplementationPrompt(feat, brainstormPath, specPath, planPath, tasksPath, summary, progress, projectRoot, outputOnly)
}

// selectFeatureForImplementation shows an interactive numbered list of
// implement-phase features.
func selectFeatureForImplementation(specsDir string) (*feature.Feature, error) {
	candidates, err := workflowStageCandidates(specsDir, workflowSelectionStageImplement)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no legacy staged features ready for implementation (need SPEC.md + PLAN.md + TASKS.md with incomplete tasks)\n\nRun 'kit legacy tasks <feature>' to create tasks first")
	}

	printSelectionHeader("Select a feature to implement:")
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

func outputImplementationPrompt(feat *feature.Feature, brainstormPath, specPath, planPath, tasksPath, summary string, progress feature.TaskProgress, projectRoot string, outputOnly bool) error {
	prompt := buildImplementationPrompt(feat, brainstormPath, specPath, planPath, tasksPath, summary, projectRoot)

	if !outputOnly {
		printImplementationContext(feat, brainstormPath, specPath, planPath, tasksPath, summary, progress)
	}

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, implementCopy); err != nil {
		return err
	}

	return nil
}
func buildImplementationPrompt(feat *feature.Feature, brainstormPath, specPath, planPath, tasksPath, summary, projectRoot string) string {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	agentDocsPath := filepath.Join(projectRoot, "docs", "agents", "README.md")
	referencesPath := filepath.Join(projectRoot, "docs", "references", "README.md")
	hasBrainstorm := document.Exists(brainstormPath)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Implement every remaining non-blocked task for legacy feature `%s`.", feat.Slug))
		if strings.TrimSpace(summary) != "" {
			doc.Paragraph(summary)
		}

		doc.Heading(2, "Context")
		rows := [][]string{
			{"TASKS.md", tasksPath, "Executable task order and status"},
			{"PLAN.md", planPath, "Implementation decisions linked by the selected task"},
			{"SPEC.md", specPath, "Binding scope and acceptance"},
			{"Agent routing", agentDocsPath, "Load only when present and relevant"},
			{"Repository references", referencesPath, "Load linked durable context only when relevant"},
			{"Constitution", constitutionPath, "Load only for project-wide invariants"},
		}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM.md", brainstormPath, "Non-binding rationale only"})
		}
		doc.Table([]string{"Input", "Path", "Use"}, rows)

		doc.Heading(2, "Execution Contract")
		doc.OrderedList(1,
			"Inspect `git status --short`; preserve unrelated and user-owned work.",
			"Open `TASKS.md`, select the next incomplete unblocked task, and load only its linked PLAN/SPEC context plus the code/tests it will touch.",
			"Run a concise implementation readiness gate: the task is unambiguous, in scope, mapped to acceptance/evidence, and compatible with current repository state. Repair discoverable doc drift first.",
			"Implement the smallest production-ready change using existing patterns. Ask only when a material choice remains non-discoverable; do not request routine approval for safe in-scope work.",
			fmt.Sprintf("Run declared checks, including `kit legacy verify %s --task <task-id>` when applicable, and record exact validation evidence.", feat.Slug),
			"Update tests, affected documentation, `TASKS.md`, and the project progress summary to match reality; mark a task done only after its acceptance and checks pass.",
			"Repeat in dependency order until every non-blocked task is complete. Do not stop after one task while safe in-scope work remains.",
		)

		doc.Heading(2, "Constraints")
		doc.BulletList(
			"Authority: safety/user constraints, Constitution, SPEC, PLAN, TASKS, then non-binding brainstorm and repo conventions.",
			"Do not invent scope, gold-plate, or add abstractions/public surfaces not required by the binding docs.",
			"Any failed or unrun validation remains explicit with reason, risk, and next action; never claim evidence that was not observed.",
			"Before issue, branch, stage, commit, push, PR, or review-thread mutation, load repo-local delivery rules and establish the exact Delivery Contract. Stop on an unknown field.",
		)

		doc.Heading(2, "Success And Output")
		doc.BulletList(
			"All non-blocked TASKS.md items and mapped acceptance criteria are complete with validation and documentation evidence.",
			"The final diff contains no known relevant defect or unrelated scope, and durable task/progress state matches it.",
			"Report outcome, files, exact checks/results, how to exercise the change, delivery state, and only genuine blockers or residual risk.",
		)
		addFinalResponseContract(doc, implementFinalResponseContract()...)
	})
}

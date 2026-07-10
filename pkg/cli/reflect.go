// package cli implements the Kit command-line interface.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/spf13/cobra"
)

var reflectCopy bool
var reflectOutputOnly bool

var reflectCmd = &cobra.Command{
	Use:   "reflect [feature]",
	Short: "Deprecated v1 staged workflow: output reflection instructions",
	Long: `Deprecated v1 staged workflow: output instructions for reflecting on recent changes to ensure
implementation correctness.

The default v2 feature workflow records validation, reflection notes,
documentation updates, delivery state, and evidence inside SPEC.md through the
kit spec supervisor prompt. Use this command only when intentionally working in
the legacy staged artifact flow.

When a feature is specified, instructions are scoped to that feature's context.
Without a feature argument, shows an interactive selection of legacy
reflect-phase features whose task checkboxes are complete and whose reflection marker is not set.
The reflection process uses git, lint, and tests to enforce a clean, working state.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReflect,
}

func init() {
	reflectCmd.Flags().BoolVar(&reflectCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	reflectCmd.Flags().BoolVar(&reflectOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(reflectCmd)
	legacyCmd.AddCommand(reflectCmd)
}

func runReflect(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := cfg.SpecsPath(projectRoot)

	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	var feat *feature.Feature

	if len(args) == 1 {
		featureRef := args[0]
		feat, err = loadFeatureWithState(specsDir, cfg, featureRef)
		if err != nil {
			return fmt.Errorf("failed to resolve feature: %w", err)
		}
	} else {
		feat, err = selectFeatureForReflect(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	wasPaused := feat.Paused
	if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
		return err
	}
	if err := updateRollupForResume(projectRoot, cfg, feat.DirName, wasPaused); err != nil {
		return err
	}
	prompt := buildReflectPrompt(projectRoot, constitutionPath, summaryPath, brainstormPath, specPath, planPath, tasksPath, feat.Slug)
	if !outputOnly {
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
		printWorkflowInstructions("reflect", []string{
			"if issues remain, return to implement",
			"if clean, mark reflection complete",
		})
	}

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, reflectCopy); err != nil {
		return err
	}

	return nil
}

// selectFeatureForReflect shows an interactive numbered list of reflect-phase
// features.
func selectFeatureForReflect(specsDir string) (*feature.Feature, error) {
	candidates, err := workflowStageCandidates(specsDir, workflowSelectionStageReflect)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no legacy staged features ready for reflection (need all TASKS.md checkboxes complete without reflection marker)\n\nRun 'kit legacy implement <feature>' until implementation tasks are complete")
	}

	printSelectionHeader("Select a feature to reflect on:")
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

// buildReflectPrompt builds the unified reflection prompt.
func buildReflectPrompt(projectRoot, constitutionPath, summaryPath, brainstormPath, specPath, planPath, tasksPath, featureSlug string) string {
	featureSlug = strings.TrimSpace(featureSlug)
	featureScoped := featureSlug != ""
	hasBrainstorm := featureScoped && document.Exists(brainstormPath)
	cfg, _ := loadRepoInstructionContext(projectRoot)
	refreshStatus, _ := calculateProjectRefreshStatus(projectRoot, cfg, time.Now().UTC())

	return renderPromptDocument(func(doc *promptdoc.Document) {
		if featureScoped {
			doc.Heading(2, fmt.Sprintf("Reflection — Feature: %s", featureSlug))
		} else {
			doc.Heading(2, "Reflection")
		}
		doc.Paragraph("Audit the integrated change, fix valid in-scope findings, and prove the result before declaring reflection complete.")

		doc.Heading(2, "Context")
		rows := [][]string{
			{"Project root", projectRoot},
			{"Constitution", constitutionPath},
			{"Project summary", summaryPath},
		}
		if featureScoped {
			if hasBrainstorm {
				rows = append(rows, []string{"BRAINSTORM.md", brainstormPath})
			}
			rows = append(rows,
				[]string{"SPEC.md", specPath},
				[]string{"PLAN.md", planPath},
				[]string{"TASKS.md", tasksPath},
			)
		}
		doc.Table([]string{"Input", "Path"}, rows)

		doc.Heading(2, "Verification evidence")
		doc.Raw(latestVerificationEvidenceStep(projectRoot, tasksPath, featureSlug))

		doc.Heading(2, "Reflection Contract")
		doc.OrderedList(1,
			"Inspect `git status --short --branch`, unstaged and staged diffs, and recent branch history. Map each changed file to its intended requirement or task.",
			"Compare the change against the binding docs and repository invariants. Confirm task/acceptance status, scope, interfaces, migration behavior, and documentation match reality.",
			"Review changed code and tests for correctness, edge cases, errors, security, concurrency, resource use, compatibility, unnecessary public surface, dead/debug code, and repository-pattern drift.",
			"Verify test quality and required evidence. Run the relevant build, lint/typecheck, tests, runtime/manual checks, generated-doc checks, and regressions; never treat an unrun check as passed.",
			"Fix every valid in-scope finding, update affected tests/docs/task state, and rerun the checks that prove the fix. Route a material requirement conflict back to planning rather than guessing.",
			"Review the final diff again and record exact checks/results, findings fixed, remaining risk, and any skipped validation with reason and impact.",
		)

		doc.Heading(2, "Completion And Delivery Boundaries")
		rules := []string{
			"Reflection completes only when no known relevant defect, unproven acceptance criterion, stale in-scope documentation, scope contamination introduced by this feature, or unresolved verifier finding remains.",
			"Preserve unrelated and user-owned work. Do not broaden reflection into cleanup unrelated to the changed scope.",
			"Before Git/GitHub delivery mutation, load repo-local delivery rules and establish the exact Delivery Contract; reflection evidence does not authorize mutation by itself.",
			projectRefreshAdvisoryStepForStatus(refreshStatus),
		}
		if featureScoped {
			rules = append(rules, "When every gate passes, append `<!-- REFLECTION_COMPLETE -->` to TASKS.md and keep the project progress summary current.")
		}
		doc.BulletList(rules...)

		doc.Heading(2, "Output")
		doc.BulletList(
			"Outcome and reflection-complete state.",
			"Changed files and findings fixed.",
			"Exact validation commands/results and verification artifact IDs.",
			"Documentation/task trace, delivery readiness, and only genuine residual risk or follow-up.",
		)
		addFinalResponseContract(doc, reflectFinalResponseContract()...)
	})
}

func latestVerificationEvidenceStep(projectRoot, tasksPath, featureSlug string) string {
	if featureSlug == "" || tasksPath == "" {
		return "Verification evidence\n- no feature-scoped verification evidence is required for generic reflection\n- if this reflection covers declared feature checks, run `kit legacy verify <feature>` first"
	}
	featureDir := filepath.Base(filepath.Dir(tasksPath))
	run, ok, err := runstore.LatestForFeature(projectRoot, featureDir)
	if err != nil {
		return fmt.Sprintf("Verification evidence\n- unable to inspect latest run evidence: %v\n- run `kit legacy verify %s` before marking reflection complete", err, featureSlug)
	}
	if !ok {
		return fmt.Sprintf("Verification evidence\n- no local verification run found for `%s`\n- run `kit legacy verify %s` and cite the resulting run ID before marking reflection complete", featureDir, featureSlug)
	}
	lines := []string{
		"Verification evidence",
		fmt.Sprintf("- latest run: `%s`", run.RunID),
		fmt.Sprintf("- status: `%s`", run.Status),
		fmt.Sprintf("- artifacts: `%s`", run.ArtifactDir),
	}
	if len(run.TaskIDs) > 0 {
		lines = append(lines, fmt.Sprintf("- tasks covered: `%s`", strings.Join(run.TaskIDs, ", ")))
	}
	if run.Status != "pass" {
		lines = append(lines, "- do not mark reflection complete until verification evidence is passing or the blocker is explicitly documented")
	}
	return strings.Join(lines, "\n")
}

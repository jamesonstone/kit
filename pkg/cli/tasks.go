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
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var tasksCopy bool
var tasksOutputOnly bool

var tasksCmd = &cobra.Command{
	Use:   "tasks [feature]",
	Short: "Deprecated v1 staged workflow: create TASKS.md",
	Long: `Deprecated v1 staged workflow: create a new task list for a feature.

The default workflow executes from an accepted native plan without requiring a
durable task checklist. Use this command only for the legacy staged artifact
flow.

Creates:
  - TASKS.md with required sections, task table, and placeholder comments

Prerequisites:
  - PLAN.md must exist (unless --force)

If no feature is specified, shows an interactive selection of features
that have SPEC.md and PLAN.md but no TASKS.md yet.

Updates PROJECT_PROGRESS_SUMMARY.md after creation.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTasks,
}

func init() {
	tasksCmd.Flags().Bool("force", false, "create missing SPEC.md and PLAN.md with headers if they don't exist")
	tasksCmd.Flags().BoolVar(&tasksCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	tasksCmd.Flags().BoolVar(&tasksOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(tasksCmd)
	legacyCmd.AddCommand(tasksCmd)
}

func runTasks(cmd *cobra.Command, args []string) error {
	tasksForce, _ := cmd.Flags().GetBool("force")
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	promptOnly := promptOnlyEnabled(cmd)

	// find project root
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	if promptOnly {
		if tasksForce {
			return fmt.Errorf("--prompt-only cannot be used with --force")
		}
		return runTasksPromptOnly(args, projectRoot, cfg, outputOnly)
	}

	var feat *feature.Feature

	if len(args) == 0 {
		// interactive mode: select from features with SPEC + PLAN but no TASKS
		feat, err = selectFeatureForTasks(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {
		// direct mode: resolve feature by name
		featureRef := args[0]
		feat, err = loadFeatureWithState(specsDir, cfg, featureRef)
		if err != nil {
			return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", featureRef, featureRef)
		}
	}

	if !outputOnly {
		fmt.Printf("📝 Creating tasks for feature: %s\n", feat.DirName)
	}

	// check prerequisites
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")

	if !document.Exists(planPath) {
		if tasksForce || cfg.AllowOutOfOrder {
			// create SPEC.md if missing
			if !document.Exists(specPath) {
				content := templates.BuildSpecV2ArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
				if err := document.Write(specPath, content); err != nil {
					return fmt.Errorf("failed to create SPEC.md: %w", err)
				}
				if !outputOnly {
					fmt.Println("  ✓ Created SPEC.md (--force)")
				}
			}
			// create PLAN.md
			content := templates.BuildPlanArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(planPath, content); err != nil {
				return fmt.Errorf("failed to create PLAN.md: %w", err)
			}
			if !outputOnly {
				fmt.Println("  ✓ Created PLAN.md (--force)")
			}
		} else {
			return fmt.Errorf("PLAN.md not found. Run 'kit legacy plan %s' first or use --force", feat.Slug)
		}
	}

	// create TASKS.md if it doesn't exist
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if !document.Exists(tasksPath) {
		content := templates.BuildTasksArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(tasksPath, content); err != nil {
			return fmt.Errorf("failed to create TASKS.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created TASKS.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ TASKS.md already exists")
	}

	wasPaused := feat.Paused
	if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
		return err
	}

	// update PROJECT_PROGRESS_SUMMARY.md
	if err := rollup.Update(projectRoot, cfg); err != nil {
		if !outputOnly {
			fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		}
	} else if !outputOnly {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	if !outputOnly {
		fmt.Printf("\n✅ Tasks for '%s' ready!\n", feat.Slug)
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
		printNumberedNextSteps([]string{
			fmt.Sprintf("Edit %s to define atomic tasks", tasksPath),
			"Link tasks to plan items using [PLAN-XX] syntax",
			"Begin implementation",
		})
	}

	return outputTasksPrompt(feat, projectRoot, cfg, outputOnly)
}

func runTasksPromptOnly(args []string, projectRoot string, cfg *config.Config, outputOnly bool) error {
	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 0 {
		feat, err = selectFeatureForTasksPromptOnly(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", args[0], args[0])
		}
	}

	if !document.Exists(filepath.Join(feat.Path, "SPEC.md")) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}
	if !document.Exists(filepath.Join(feat.Path, "PLAN.md")) {
		return fmt.Errorf("PLAN.md not found. Run 'kit legacy plan %s' first", feat.Slug)
	}
	if !document.Exists(filepath.Join(feat.Path, "TASKS.md")) {
		return fmt.Errorf("TASKS.md not found. Run 'kit legacy tasks %s' first", feat.Slug)
	}

	return outputTasksPrompt(feat, projectRoot, cfg, outputOnly)
}

func outputTasksPrompt(feat *feature.Feature, projectRoot string, cfg *config.Config, outputOnly bool) error {
	prompt := buildTasksPrompt(feat, projectRoot, cfg)

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, tasksCopy); err != nil {
		return err
	}

	return nil
}
func buildTasksPrompt(feat *feature.Feature, projectRoot string, cfg *config.Config) string {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	hasBrainstorm := document.Exists(brainstormPath)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Create the executable legacy task plan for feature `%s`. Update TASKS.md only; do not implement product code.", feat.Slug))
		doc.Heading(2, "Context")
		rows := [][]string{
			{"SPEC.md", specPath, "Binding requirements and acceptance"},
			{"PLAN.md", planPath, "Binding implementation strategy"},
			{"TASKS.md", tasksPath, "Artifact to update"},
			{"Constitution", constitutionPath, "Project invariants"},
		}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM.md", brainstormPath, "Non-binding rationale only"})
		}
		doc.Table([]string{"Input", "Path", "Use"}, rows)

		doc.Heading(2, "Task-Generation Contract")
		doc.OrderedList(1,
			fmt.Sprintf("Update TASKS.md directly at %s; do not leave the task breakdown only in chat.", tasksPath),
			"Read SPEC.md and PLAN.md as fixed inputs. Resolve repository-discoverable ordering, file, and validation facts yourself.",
			"Ask only when a material non-discoverable decision prevents an executable task breakdown; use concise numbered questions with recommended defaults and impact, then stop until answered.",
			"Update TASKS.md directly with stable T001-style IDs, dependency order, and the smallest coherent tasks that can be completed and verified independently.",
			"Map each task to PLAN/SPEC acceptance. Every task detail includes GOAL, SCOPE, ACCEPTANCE, VERIFY, EXPECTED FILES, RISK, ROLLBACK, DEPENDENCIES, and NOTES when needed.",
			"Keep the progress table and markdown checkbox list consistent with task details. Initial implementation status is todo unless existing evidence proves otherwise.",
		)

		doc.Heading(2, "Success Criteria")
		doc.BulletList(
			fmt.Sprintf("Confidence is at least %d and no blocking task-definition question remains.", cfg.GoalPercentage),
			"Every acceptance criterion is covered; every task has a binary done condition and required evidence.",
			"Ordering and dependencies are explicit, file overlap is visible, and a coding agent can execute the list without routine back-and-forth.",
			"No requirement, implementation detail, or scope is invented beyond SPEC.md and PLAN.md; placeholders are removed and project progress is current.",
		)

		doc.Heading(2, "Output")
		doc.BulletList(
			"Update TASKS.md only.",
			"Report task count/order, evidence coverage, blockers, and the next legacy implementation step.",
		)
		addFinalResponseContract(doc, tasksFinalResponseContract(feat.Slug)...)
	})
}

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

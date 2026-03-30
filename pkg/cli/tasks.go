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
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var tasksCopy bool
var tasksOutputOnly bool

var tasksCmd = &cobra.Command{
	Use:   "tasks [feature]",
	Short: "Create or open feature tasks",
	Long: `Create a new task list for a feature.

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
	rootCmd.AddCommand(tasksCmd)
}

func runTasks(cmd *cobra.Command, args []string) error {
	tasksForce, _ := cmd.Flags().GetBool("force")
	outputOnly, _ := cmd.Flags().GetBool("output-only")

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

	var feat *feature.Feature

	if len(args) == 0 {
		// interactive mode: select from features with SPEC + PLAN but no TASKS
		feat, err = selectFeatureForTasks(specsDir)
		if err != nil {
			return err
		}
	} else {
		// direct mode: resolve feature by name
		featureRef := args[0]
		feat, err = feature.Resolve(specsDir, featureRef)
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
				if err := document.Write(specPath, templates.Spec); err != nil {
					return fmt.Errorf("failed to create SPEC.md: %w", err)
				}
				if !outputOnly {
					fmt.Println("  ✓ Created SPEC.md (--force)")
				}
			}
			// create PLAN.md
			if err := document.Write(planPath, templates.Plan); err != nil {
				return fmt.Errorf("failed to create PLAN.md: %w", err)
			}
			if !outputOnly {
				fmt.Println("  ✓ Created PLAN.md (--force)")
			}
		} else {
			return fmt.Errorf("PLAN.md not found. Run 'kit plan %s' first or use --force", feat.Slug)
		}
	}

	// create TASKS.md if it doesn't exist
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if !document.Exists(tasksPath) {
		if err := document.Write(tasksPath, templates.Tasks); err != nil {
			return fmt.Errorf("failed to create TASKS.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created TASKS.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ TASKS.md already exists")
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
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit %s to define atomic tasks\n", tasksPath)
		fmt.Printf("  2. Link tasks to plan items using [PLAN-XX] syntax\n")
		fmt.Printf("  3. Begin implementation!\n")
	}

	// output easy-to-copy instruction for coding agents
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	hasBrainstorm := document.Exists(brainstormPath)
	goalPct := cfg.GoalPercentage

	var sb strings.Builder
	sb.WriteString("Please review and complete the task plan.\n\n")
	sb.WriteString("## File References\n")
	sb.WriteString("| Document | Path |\n")
	sb.WriteString("|----------|------|\n")
	sb.WriteString(fmt.Sprintf("| CONSTITUTION | %s |\n", constitutionPath))
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("| BRAINSTORM | %s |\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("| SPEC | %s |\n", specPath))
	sb.WriteString(fmt.Sprintf("| PLAN | %s |\n", planPath))
	sb.WriteString(fmt.Sprintf("| TASKS | %s |\n", tasksPath))
	sb.WriteString(fmt.Sprintf("| Feature | %s |\n", feat.Slug))
	sb.WriteString(fmt.Sprintf("| Project Root | %s |\n\n", projectRoot))

	sb.WriteString("Your task:\n")
	sb.WriteString(fmt.Sprintf("1. Read CONSTITUTION.md (file: %s) to understand project constraints and principles\n", constitutionPath))
	step := 2
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("%d. Read BRAINSTORM.md (file: %s) to preserve upstream research context\n", step, brainstormPath))
		step++
	}
	sb.WriteString(fmt.Sprintf("%d. Read SPEC.md (file: %s) and PLAN.md (file: %s) fully and treat them as fixed inputs\n", step, specPath, planPath))
	step++
	sb.WriteString(fmt.Sprintf("%d. Review the TASKS.md (file: %s) template and required sections\n", step, tasksPath))
	step++
	sb.WriteString(fmt.Sprintf("%d. Derive an atomic, ordered task list that can be executed without ambiguity\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. Identify missing decisions that block task generation\n", step))
	step++
	appendNumberedSteps(
		&sb,
		step,
		clarificationLoopSteps(
			goalPct,
			"Reassess and continue with additional batches of up to 10 questions "+
				"until the task plan is precise enough to produce a correct, "+
				"production-quality implementation",
		),
	)
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf(`Before you write or update TASKS.md:
- after each batch of up to 10 questions, output your current percentage understanding of the task plan so the user can see progress
- do NOT treat TASKS.md as complete until confidence reaches ≥%d%%

TASKS.md requirements:

A) PROGRESS TABLE (ALWAYS FIRST)
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

E) NOTES SECTION
- only if required to prevent ambiguity

F) PLAN LINKS (OPTIONAL)
- Link to PLAN sections using anchors from headings (lowercase, dashes)
- Examples: [PLAN-APPROACH], [PLAN-COMPONENTS], [PLAN-DATA], [PLAN-RISKS]
- Include in task descriptions to trace back to plan requirements

Rules:
- focus on executable steps, not prose
- use BRAINSTORM.md as research context only; SPEC.md and PLAN.md remain the binding inputs
- do not invent new requirements or scope beyond SPEC.md
- tasks must map back to PLAN items via section anchors
- tasks must imply an unambiguous implementation order
- Tasks gate: each task must include an explicit done-condition and required evidence artifact(s) before sign-off
- avoid code unless strictly necessary
- keep language dense and factual
- ensure tasks respect constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

Output goal:
- a task list that a coding agent can execute linearly with minimal back-and-forth
`, goalPct))

	prompt := sb.String()

	if err := outputPromptWithClipboardDefault(prompt, outputOnly, tasksCopy); err != nil {
		return err
	}

	return nil
}

// selectFeatureForTasks shows an interactive numbered list of features
// that have SPEC.md and PLAN.md but no TASKS.md yet.
func selectFeatureForTasks(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	// filter to features with SPEC + PLAN but no TASKS
	var candidates []feature.Feature
	for _, f := range features {
		specPath := filepath.Join(f.Path, "SPEC.md")
		planPath := filepath.Join(f.Path, "PLAN.md")
		tasksPath := filepath.Join(f.Path, "TASKS.md")
		if document.Exists(specPath) && document.Exists(planPath) && !document.Exists(tasksPath) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features ready for tasks (need SPEC.md + PLAN.md without TASKS.md)\n\nRun 'kit plan <feature>' to create a plan first")
	}

	fmt.Println()
	fmt.Println(whiteBold + "Select a feature to create tasks for:" + reset)
	fmt.Println()
	for i, f := range candidates {
		fmt.Printf("  [%d] %s\n", i+1, f.DirName)
	}
	fmt.Println()
	fmt.Print(whiteBold + "Enter number: " + reset)

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

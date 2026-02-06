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
)

var implementCopy bool

var implementCmd = &cobra.Command{
	Use:   "implement [feature]",
	Short: "Output implementation context for coding agents",
	Long: `Output a comprehensive summary for coding agents to begin implementation.

Provides:
  - Feature overview and current status
  - Document reference table (SPEC, PLAN, TASKS)
  - Clear instructions for executing tasks

If no feature is specified, shows an interactive selection of features
that have SPEC.md, PLAN.md, and TASKS.md ready for implementation.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runImplement,
}

func init() {
	implementCmd.Flags().BoolVar(&implementCopy, "copy", false, "copy agent prompt to clipboard (suppresses stdout)")
	rootCmd.AddCommand(implementCmd)
}

func runImplement(cmd *cobra.Command, args []string) error {
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
	} else {
		// direct mode: resolve feature by name
		featureRef := args[0]
		feat, err = feature.Resolve(specsDir, featureRef)
		if err != nil {
			return fmt.Errorf("feature '%s' not found", featureRef)
		}
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	// verify all documents exist
	if !document.Exists(specPath) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}
	if !document.Exists(planPath) {
		return fmt.Errorf("PLAN.md not found. Run 'kit plan %s' first", feat.Slug)
	}
	if !document.Exists(tasksPath) {
		return fmt.Errorf("TASKS.md not found. Run 'kit tasks %s' first", feat.Slug)
	}

	// extract summary from spec
	summary, _ := feature.ExtractSpecSummary(specPath)

	// get task progress
	progress, _ := feature.ParseTaskProgress(tasksPath)

	return outputImplementationPrompt(feat, specPath, planPath, tasksPath, summary, progress, projectRoot)
}

// selectFeatureForImplementation shows an interactive numbered list of features
// that have SPEC.md, PLAN.md, and TASKS.md ready.
func selectFeatureForImplementation(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	// filter to features with all three documents
	var candidates []feature.Feature
	for _, f := range features {
		specPath := filepath.Join(f.Path, "SPEC.md")
		planPath := filepath.Join(f.Path, "PLAN.md")
		tasksPath := filepath.Join(f.Path, "TASKS.md")
		if document.Exists(specPath) && document.Exists(planPath) && document.Exists(tasksPath) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features ready for implementation (need SPEC.md + PLAN.md + TASKS.md)\n\nRun 'kit tasks <feature>' to create tasks first")
	}

	fmt.Println()
	fmt.Println(whiteBold + "Select a feature to implement:" + reset)
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

func outputImplementationPrompt(feat *feature.Feature, specPath, planPath, tasksPath, summary string, progress feature.TaskProgress, projectRoot string) error {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")

	// build the agent prompt
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are implementing the feature: %s\n\n## Overview\n", feat.Slug))

	if summary != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", summary))
	} else {
		sb.WriteString("(Read SPEC.md for feature description)\n\n")
	}

	sb.WriteString(fmt.Sprintf(`## Document Hierarchy

| Document | Contains | Use When |
|----------|----------|----------|
| CONSTITUTION.md | Project-wide constraints, principles, priors | Understanding fundamental rules |
| SPEC.md | Requirements, goals, constraints, acceptance criteria | Checking scope, validating completeness |
| PLAN.md | Architecture, components, interfaces, design decisions | Making implementation choices, understanding structure |
| TASKS.md | Ordered execution steps with acceptance criteria per task | Knowing what to do next, tracking progress |

## Your Instructions

1. **Read CONSTITUTION.md first** to understand project constraints and principles
2. **Read all three feature documents** in order: SPEC → PLAN → TASKS
3. **Supplement with your context**: If you have internal plans, prior conversation context, or a Warp plan related to this feature, use that knowledge to inform your implementation — but always defer to CONSTITUTION/SPEC/PLAN/TASKS when there's a conflict
4. **Execute tasks from TASKS.md** in the order specified
5. **For each task:**
   - Read the task's GOAL, SCOPE, and ACCEPTANCE criteria
   - Implement only what's specified (no gold-plating)
   - Verify acceptance criteria are met before marking complete
   - Update TASKS.md: change '- [ ]' to '- [x]' when done

## Key Files
- CONSTITUTION: %s
- SPEC: %s
- PLAN: %s
- TASKS: %s
- Project root: %s

## Rules
- Respect constraints defined in CONSTITUTION.md
- Stay within scope defined in SPEC.md
- Follow architecture decisions in PLAN.md
- Complete tasks in dependency order from TASKS.md
- Ask for clarification rather than making assumptions
- If a task is blocked, explain what's blocking and suggest resolution
- After completing each task, briefly confirm what was done
- **Use available tools**: If you have access to MCP servers (e.g., Context7 for documentation, GitHub for issues/PRs, or others), use them to fetch up-to-date documentation, verify API usage, and ensure implementation correctness
- **Always** update %s/docs/PROJECT_PROGRESS_SUMMARY.md as progress is made and at implementation completion
- Keep TASKS.md updated with accurate status and ensure that it reflects reality upon completion

## Begin
Start by reading TASKS.md to identify the first incomplete task (marked with '- [ ]').
Then read its acceptance criteria and implement it.
`, constitutionPath, specPath, planPath, tasksPath, projectRoot, projectRoot))

	prompt := sb.String()

	// copy to clipboard if requested
	if implementCopy {
		if err := copyToClipboard(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✓ Copied agent prompt to clipboard")
		return nil
	}

	fmt.Println()
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "🚀 Implementation Context: " + reset + feat.DirName)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()

	// feature summary
	if summary != "" {
		fmt.Println(whiteBold + "Feature Summary:" + reset)
		fmt.Println(summary)
		fmt.Println()
	}

	// progress status
	if progress.HasTasks() {
		fmt.Printf(whiteBold+"Progress: "+reset+"%d/%d tasks complete\n", progress.Complete, progress.Total)
	} else {
		fmt.Println(whiteBold + "Progress: " + reset + "Tasks defined, ready to begin")
	}
	fmt.Println()

	// document reference table
	fmt.Println(whiteBold + "Document Reference:" + reset)
	fmt.Println(dim + "┌─────────────┬────────────────────────────────────────────────────────────┐" + reset)
	fmt.Println(dim + "│ " + reset + whiteBold + "Document" + reset + dim + "    │ " + reset + whiteBold + "Purpose & Usage" + reset + dim + "                                          │" + reset)
	fmt.Println(dim + "├─────────────┼────────────────────────────────────────────────────────────┤" + reset)
	fmt.Println(dim + "│ " + reset + "SPEC.md" + dim + "     │ " + reset + "WHAT: Requirements, constraints, acceptance criteria" + dim + "      │" + reset)
	fmt.Println(dim + "│             │ " + reset + "→ Consult when unsure if something is in scope" + dim + "            │" + reset)
	fmt.Println(dim + "│             │ " + reset + "→ Do NOT add features not specified here" + dim + "                  │" + reset)
	fmt.Println(dim + "├─────────────┼────────────────────────────────────────────────────────────┤" + reset)
	fmt.Println(dim + "│ " + reset + "PLAN.md" + dim + "     │ " + reset + "HOW: Architecture, components, data structures" + dim + "           │" + reset)
	fmt.Println(dim + "│             │ " + reset + "→ Follow the design decisions made here" + dim + "                   │" + reset)
	fmt.Println(dim + "│             │ " + reset + "→ If blocked, check RISKS section for mitigations" + dim + "         │" + reset)
	fmt.Println(dim + "├─────────────┼────────────────────────────────────────────────────────────┤" + reset)
	fmt.Println(dim + "│ " + reset + "TASKS.md" + dim + "    │ " + reset + "EXECUTE: Ordered task list with acceptance criteria" + dim + "       │" + reset)
	fmt.Println(dim + "│             │ " + reset + "→ Work through tasks in order (respect dependencies)" + dim + "       │" + reset)
	fmt.Println(dim + "│             │ " + reset + "→ Mark tasks complete with [x] when acceptance met" + dim + "         │" + reset)
	fmt.Println(dim + "└─────────────┴────────────────────────────────────────────────────────────┘" + reset)
	fmt.Println()

	// file paths
	fmt.Println(whiteBold + "File Locations:" + reset)
	fmt.Printf("  • SPEC:  %s\n", specPath)
	fmt.Printf("  • PLAN:  %s\n", planPath)
	fmt.Printf("  • TASKS: %s\n", tasksPath)
	fmt.Println()

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Print(prompt)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	return nil
}

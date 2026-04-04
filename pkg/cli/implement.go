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
var implementOutputOnly bool

var implementCmd = &cobra.Command{
	Use:   "implement [feature]",
	Short: "Run the readiness gate and output implementation context",
	Long: `Run the implementation readiness gate and output a comprehensive
summary for coding agents to begin implementation.

Provides:
  - Implementation-readiness gate instructions
  - Feature overview and current status
  - Document reference table (SPEC, PLAN, TASKS)
  - Clear instructions for executing tasks

If no feature is specified, shows an interactive selection of features
that have SPEC.md, PLAN.md, and TASKS.md ready for implementation.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runImplement,
}

func init() {
	implementCmd.Flags().BoolVar(&implementCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	implementCmd.Flags().BoolVar(&implementOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(implementCmd)
	rootCmd.AddCommand(implementCmd)
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
		return fmt.Errorf("PLAN.md not found. Run 'kit plan %s' first", feat.Slug)
	}
	if !document.Exists(tasksPath) {
		return fmt.Errorf("TASKS.md not found. Run 'kit tasks %s' first", feat.Slug)
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

	if err := outputPromptWithClipboardDefault(prompt, outputOnly, implementCopy); err != nil {
		return err
	}

	return nil
}

func buildImplementationPrompt(feat *feature.Feature, brainstormPath, specPath, planPath, tasksPath, summary, projectRoot string) string {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	hasBrainstorm := document.Exists(brainstormPath)

	// build the agent prompt
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are implementing the feature: %s\n\n## Overview\n", feat.Slug))

	if summary != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", summary))
	} else {
		sb.WriteString("(Read SPEC.md for feature description)\n\n")
	}

	sb.WriteString("## Document Hierarchy\n\n")
	sb.WriteString("| Document | Contains | Use When |\n")
	sb.WriteString("|----------|----------|----------|\n")
	sb.WriteString("| CONSTITUTION.md | Project-wide constraints, principles, priors | Understanding fundamental rules |\n")
	if hasBrainstorm {
		sb.WriteString("| BRAINSTORM.md | Upstream research findings, relevant files, strategy options | Recovering problem context before execution |\n")
	}
	sb.WriteString("| SPEC.md | Requirements, goals, constraints, acceptance criteria | Checking scope, validating completeness |\n")
	sb.WriteString("| PLAN.md | Architecture, components, interfaces, design decisions | Making implementation choices, understanding structure |\n")
	sb.WriteString("| TASKS.md | Ordered execution steps with acceptance criteria per task | Knowing what to do next, tracking progress |\n\n")

	sb.WriteString("## Your Instructions\n\n")
	sb.WriteString("1. **Read CONSTITUTION.md first** to understand project constraints and principles\n")
	sb.WriteString("2. **Read the feature documents in order**: ")
	if hasBrainstorm {
		sb.WriteString("BRAINSTORM → ")
	}
	sb.WriteString("SPEC → PLAN → TASKS\n")
	sb.WriteString("3. **Run the implementation readiness gate before writing code**\n")
	sb.WriteString(`   - Treat this as an adversarial preflight over the full document set
   - Build a review map for CONSTITUTION.md, optional BRAINSTORM.md, SPEC.md, PLAN.md, and TASKS.md
   - Challenge the docs for contradictions, ambiguous requirements, hidden assumptions, missing edge cases or failure modes, task gaps, and scope creep
   - Verify each planned implementation step still traces back to the binding docs
   - Produce an explicit go/no-go decision before coding

`)
	sb.WriteString(`4. **If the readiness gate fails, stop and repair the docs first**
   - Do NOT write production code yet
   - Update SPEC.md, PLAN.md, and/or TASKS.md to resolve the exact issue
   - Update PROJECT_PROGRESS_SUMMARY.md when the feature summary or state changes
   - Re-run the implementation readiness gate after the docs are fixed

`)
	sb.WriteString("5. **Supplement with your context**: If you have internal plans, prior conversation context, or a Warp plan related to this feature, use that knowledge to inform your implementation — but always defer to CONSTITUTION/SPEC/PLAN/TASKS when there's a conflict\n")
	sb.WriteString("6. **Only after the readiness gate passes, execute tasks from TASKS.md**\n")
	sb.WriteString(`7. **For each task:**
   - Start with the first incomplete task (marked '- [ ]')
   - Read the task's GOAL, SCOPE, and ACCEPTANCE criteria
   - Implement only what's specified (no gold-plating)
   - Verify acceptance criteria are met before marking complete
   - Update TASKS.md: change '- [ ]' to '- [x]' when done

`)

	sb.WriteString("## Key Files\n")
	sb.WriteString(fmt.Sprintf("- CONSTITUTION: %s\n", constitutionPath))
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("- BRAINSTORM: %s\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("- SPEC: %s\n", specPath))
	sb.WriteString(fmt.Sprintf("- PLAN: %s\n", planPath))
	sb.WriteString(fmt.Sprintf("- TASKS: %s\n", tasksPath))
	sb.WriteString(fmt.Sprintf("- Project root: %s\n\n", projectRoot))

	sb.WriteString(fmt.Sprintf(`## Rules
- Respect constraints defined in CONSTITUTION.md
- Use BRAINSTORM.md as context only; SPEC, PLAN, and TASKS control execution
- Do not begin coding until the implementation readiness gate passes
- If the readiness gate fails, report the exact contradiction, ambiguity, missing coverage, or scope issue before editing docs
- Re-run the readiness gate every time implementation restarts after a doc repair
- Stay within scope defined in SPEC.md
- Follow architecture decisions in PLAN.md
- Complete tasks in dependency order from TASKS.md
- Quality gate: target zero known defects; do not mark implementation complete until all gates pass with evidence: unresolved assumptions = 0, acceptance criteria mapped 1:1 to outputs, build/compile succeeds, lint/typecheck/test failures = 0, and unrelated diff scope = 0
- If any gate fails, stop, report the exact failure, and propose the next fix
- Ask for clarification rather than making assumptions
- If a task is blocked, explain what's blocking and suggest resolution
- After completing each task, briefly confirm what was done
- **Use available tools**: If you have access to MCP servers (e.g., Context7 for documentation, GitHub for issues/PRs, or others), use them to fetch up-to-date documentation, verify API usage, and ensure implementation correctness
- **Always** update %s/docs/PROJECT_PROGRESS_SUMMARY.md as progress is made and at implementation completion
- Keep TASKS.md updated with accurate status and ensure that it reflects reality upon completion

## Begin
Start by running the implementation readiness gate against the document set.
Do not write code until the gate passes.
Once it passes, read TASKS.md to identify the first incomplete task (marked with '- [ ]').
Then read its acceptance criteria and implement it.
`, projectRoot))

	return sb.String()
}

func printImplementationContext(feat *feature.Feature, brainstormPath, specPath, planPath, tasksPath, summary string, progress feature.TaskProgress) {
	hasBrainstorm := document.Exists(brainstormPath)
	style := styleForStdout()

	fmt.Println()
	fmt.Println(style.title("🛠️", fmt.Sprintf("Implementation Context: %s", feat.DirName)))
	fmt.Println()

	if summary != "" {
		fmt.Println(style.title("📝", "Feature Summary"))
		fmt.Println(summary)
		fmt.Println()
	} else {
		fmt.Println(style.title("📝", "Feature Summary"))
		fmt.Println("(Read SPEC.md for feature description)")
		fmt.Println()
	}

	if progress.HasTasks() {
		fmt.Println(style.title("📈", fmt.Sprintf("Progress: %d/%d tasks complete", progress.Complete, progress.Total)))
	} else {
		fmt.Println(style.title("📈", "Progress: Tasks defined, ready to begin"))
	}
	fmt.Println()

	fmt.Println(style.title("📚", "Document Reference"))
	printImplementDocumentReferenceTable()
	fmt.Println()

	fmt.Println(style.title("📍", "File Locations"))
	if hasBrainstorm {
		fmt.Printf("  • BRAINSTORM: %s\n", brainstormPath)
	}
	fmt.Printf("  • SPEC:  %s\n", specPath)
	fmt.Printf("  • PLAN:  %s\n", planPath)
	fmt.Printf("  • TASKS: %s\n", tasksPath)
	fmt.Println()
}

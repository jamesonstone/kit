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
	cfg, _ := loadRepoInstructionContext(projectRoot)
	repoAgentsPath := repoKnowledgeEntrypointPath(projectRoot, cfg)
	repoReferencesPath := repoReferencesEntrypointPath(projectRoot, cfg)

	steps := []string{}
	if repoAgentsPath != "" {
		steps = append(steps, "**If present, read `docs/agents/README.md`** and only the linked workflow docs relevant to this feature")
	}
	if repoReferencesPath != "" {
		steps = append(steps, "**If a repo-wide reference matters, read `docs/references/README.md`** and only the linked reference docs relevant to this feature")
	}
	documentOrder := "SPEC → PLAN → TASKS"
	if hasBrainstorm {
		documentOrder = "BRAINSTORM → " + documentOrder
	}
	steps = append(steps,
		"**Read CONSTITUTION.md first** to understand project constraints and principles",
		fmt.Sprintf("**Read the feature documents in order**: %s", documentOrder),
		"**Run the implementation readiness gate before writing code**\n- Treat this as an adversarial preflight over the full document set\n- Build a review map for CONSTITUTION.md, optional BRAINSTORM.md, SPEC.md, PLAN.md, and TASKS.md\n- Challenge the docs for contradictions, ambiguous requirements, hidden assumptions, missing edge cases or failure modes, task gaps, and scope creep\n- Verify each planned implementation step still traces back to the binding docs\n- Produce an explicit go/no-go decision before coding",
		"**If the readiness gate fails, stop and repair the docs first**\n- Do NOT write production code yet\n- Update SPEC.md, PLAN.md, and/or TASKS.md to resolve the exact issue\n- Update PROJECT_PROGRESS_SUMMARY.md when the feature summary or state changes\n- Re-run the implementation readiness gate after the docs are fixed",
		"**Supplement with your context**: If you have internal plans, prior conversation context, or a Warp plan related to this feature, use that knowledge to inform your implementation — but always defer to CONSTITUTION/SPEC/PLAN/TASKS when there's a conflict",
		"**Only after the readiness gate passes, execute tasks from TASKS.md**",
		"**For each task:**\n- Start with the first incomplete task (marked '- [ ]')\n- Read the task's GOAL, SCOPE, and ACCEPTANCE criteria\n- Implement only what's specified (no gold-plating)\n- Verify acceptance criteria are met before marking complete\n- Update TASKS.md: change '- [ ]' to '- [x]' when done",
	)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("You are implementing the feature: %s", feat.Slug))
		doc.Heading(2, "Overview")
		if summary != "" {
			doc.Paragraph(summary)
		} else {
			doc.Paragraph("(Read SPEC.md for feature description)")
		}
		doc.Heading(2, "Document Hierarchy")
		rows := [][]string{}
		if repoAgentsPath != "" {
			rows = append(rows, []string{
				"docs/agents/README.md",
				"Repo-local workflow index, guardrails, and RLM routing",
				"Reconstructing the thin ToC model before feature execution",
			})
		}
		if repoReferencesPath != "" {
			rows = append(rows, []string{
				"docs/references/README.md",
				"Repo-wide references index for durable background context",
				"A repo-wide reference materially shapes the feature",
			})
		}
		rows = append(rows, []string{
			"CONSTITUTION.md",
			"Project-wide constraints, principles, priors",
			"Understanding fundamental rules",
		})
		if hasBrainstorm {
			rows = append(rows, []string{
				"BRAINSTORM.md",
				"Upstream research findings, relevant files, strategy options",
				"Recovering problem context before execution",
			})
		}
		rows = append(rows,
			[]string{"SPEC.md", "Requirements, goals, constraints, acceptance criteria", "Checking scope, validating completeness"},
			[]string{"PLAN.md", "Architecture, components, interfaces, design decisions", "Making implementation choices, understanding structure"},
			[]string{"TASKS.md", "Ordered execution steps with acceptance criteria per task", "Knowing what to do next, tracking progress"},
		)
		doc.Table([]string{"Document", "Contains", "Use When"}, rows)
		doc.Heading(2, "Your Instructions")
		doc.OrderedList(1, steps...)
		doc.Heading(2, "Key Files")
		keyFiles := []string{}
		if repoAgentsPath != "" {
			keyFiles = append(keyFiles, fmt.Sprintf("AGENTS DOCS: %s", repoAgentsPath))
		}
		if repoReferencesPath != "" {
			keyFiles = append(keyFiles, fmt.Sprintf("REFERENCES: %s", repoReferencesPath))
		}
		keyFiles = append(keyFiles, fmt.Sprintf("CONSTITUTION: %s", constitutionPath))
		if hasBrainstorm {
			keyFiles = append(keyFiles, fmt.Sprintf("BRAINSTORM: %s", brainstormPath))
		}
		keyFiles = append(keyFiles,
			fmt.Sprintf("SPEC: %s", specPath),
			fmt.Sprintf("PLAN: %s", planPath),
			fmt.Sprintf("TASKS: %s", tasksPath),
			fmt.Sprintf("Project root: %s", projectRoot),
		)
		doc.BulletList(keyFiles...)
		doc.Heading(2, "Rules")
		doc.BulletList(
			"Respect constraints defined in CONSTITUTION.md",
			"Use BRAINSTORM.md as context only; SPEC, PLAN, and TASKS control execution",
			"Do not begin coding until the implementation readiness gate passes",
			"If the readiness gate fails, report the exact contradiction, ambiguity, missing coverage, or scope issue before editing docs",
			"Re-run the readiness gate every time implementation restarts after a doc repair",
			"Stay within scope defined in SPEC.md",
			"Follow architecture decisions in PLAN.md",
			"Complete tasks in dependency order from TASKS.md",
			"Quality gate: target zero known defects; do not mark implementation complete until all gates pass with evidence: unresolved assumptions = 0, acceptance criteria mapped 1:1 to outputs, build/compile succeeds, lint/typecheck/test failures = 0, and unrelated diff scope = 0",
			"If any gate fails, stop, report the exact failure, and propose the next fix",
			"Ask for clarification rather than making assumptions",
			"If a task is blocked, explain what's blocking and suggest resolution",
			"After completing each task, briefly confirm what was done",
			"**Use available tools**: If you have access to MCP servers (e.g., Context7 for documentation, GitHub for issues/PRs, or others), use them to fetch up-to-date documentation, verify API usage, and ensure implementation correctness",
			fmt.Sprintf("**Always** update %s/docs/PROJECT_PROGRESS_SUMMARY.md as progress is made and at implementation completion", projectRoot),
			"Keep TASKS.md updated with accurate status and ensure that it reflects reality upon completion",
		)
		doc.Heading(2, "Begin")
		doc.Paragraph("Start by running the implementation readiness gate against the document set.")
		doc.Paragraph("Do not write code until the gate passes.")
		doc.Paragraph("Once it passes, read TASKS.md to identify the first incomplete task (marked with '- [ ]').")
		doc.Paragraph("Then read its acceptance criteria and implement it.")
	})
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

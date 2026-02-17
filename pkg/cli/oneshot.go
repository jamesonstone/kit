package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/git"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var (
	oneshotCopy     bool
	oneshotSpec     string
	oneshotSpecFile string
)

var oneshotCmd = &cobra.Command{
	Use:   "oneshot <feature>",
	Short: "Scaffold all feature artifacts and output a combined agent prompt",
	Long: `Create all feature artifacts (SPEC.md, PLAN.md, TASKS.md) in one step
and output a comprehensive prompt that drives a coding agent through the
entire spec-driven workflow.

The agent will:
  1. Read your brainstorming specification
  2. Drive clarification until >= 95% understanding
  3. Fill out SPEC.md, PLAN.md, and TASKS.md progressively
  4. Reach a pre-implementation phase ready for execution

Modes:
  Default:     Prompts you to paste the brainstorming specification
  --spec:      Pass the brainstorming specification inline
  --spec-file: Read the brainstorming specification from a file

Examples:
  kit oneshot my-feature
  kit oneshot my-feature --spec "Add CSV export with streaming support"
  kit oneshot my-feature --spec-file docs/brainstorm-export.md`,
	Args: cobra.ExactArgs(1),
	RunE: runOneshot,
}

func init() {
	oneshotCmd.Flags().BoolVar(&oneshotCopy, "copy", false, "copy agent prompt to clipboard")
	oneshotCmd.Flags().StringVar(&oneshotSpec, "spec", "", "brainstorming specification text (inline)")
	oneshotCmd.Flags().StringVar(&oneshotSpecFile, "spec-file", "", "path to brainstorming specification file")
	oneshotCmd.Flags().Bool("create-branch", false, "create and switch to a git branch matching the feature name")
	rootCmd.AddCommand(oneshotCmd)
}

func runOneshot(cmd *cobra.Command, args []string) error {
	createBranch, _ := cmd.Flags().GetBool("create-branch")
	featureRef := args[0]

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)
	if err := ensureDir(specsDir); err != nil {
		return err
	}

	// resolve brainstorming specification
	brainstormText, err := resolveBrainstormSpec(oneshotSpec, oneshotSpecFile)
	if err != nil {
		return err
	}

	// create or find feature
	feat, created, err := feature.EnsureExists(cfg, specsDir, featureRef)
	if err != nil {
		return err
	}

	if created {
		fmt.Printf("📁 Created feature directory: %s\n", feat.DirName)
	} else {
		fmt.Printf("📁 Using existing feature: %s\n", feat.DirName)
	}

	// create all artifact files
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	if err := ensureArtifact(specPath, templates.Spec, "SPEC.md"); err != nil {
		return err
	}
	if err := ensureArtifact(planPath, templates.Plan, "PLAN.md"); err != nil {
		return err
	}
	if err := ensureArtifact(tasksPath, templates.Tasks, "TASKS.md"); err != nil {
		return err
	}

	// create git branch if requested
	if createBranch && git.IsRepo(projectRoot) {
		createBranchForFeature(projectRoot, feat, cfg)
	}

	// update PROJECT_PROGRESS_SUMMARY.md
	if err := rollup.Update(projectRoot, cfg); err != nil {
		fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
	} else {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	fmt.Printf("\n✅ Feature '%s' fully scaffolded!\n", feat.Slug)

	return outputOneshotPrompt(feat, specPath, planPath, tasksPath, brainstormText, projectRoot, cfg)
}

// resolveBrainstormSpec gets the brainstorming spec from flag, file, or interactive input.
func resolveBrainstormSpec(inline, filePath string) (string, error) {
	if inline != "" && filePath != "" {
		return "", fmt.Errorf("cannot use both --spec and --spec-file")
	}

	if inline != "" {
		return inline, nil
	}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read spec file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	return readBrainstormInteractive()
}

// readBrainstormInteractive reads a multi-line brainstorming spec from stdin.
func readBrainstormInteractive() (string, error) {
	fmt.Println()
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "📝 Paste your brainstorming specification" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(dim + "Paste or type your feature description, brainstorm, or rough spec." + reset)
	fmt.Println(dim + "Type '===END===' on its own line or press Ctrl+D (EOF) when done." + reset)
	fmt.Println()

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "===END===" {
			break
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	result := strings.TrimSpace(strings.Join(lines, "\n"))
	if result == "" {
		return "", fmt.Errorf("no brainstorming specification provided")
	}

	fmt.Printf("\n  ✓ Received %d lines of brainstorming specification\n", len(lines))
	return result, nil
}

// ensureArtifact creates a document file if it doesn't already exist.
func ensureArtifact(path, template, name string) error {
	if !document.Exists(path) {
		if err := document.Write(path, template); err != nil {
			return fmt.Errorf("failed to create %s: %w", name, err)
		}
		fmt.Printf("  ✓ Created %s\n", name)
	} else {
		fmt.Printf("  ✓ %s already exists\n", name)
	}
	return nil
}

// outputOneshotPrompt generates the combined 5-phase agent prompt.
func outputOneshotPrompt(feat *feature.Feature, specPath, planPath, tasksPath, brainstormText, projectRoot string, cfg *config.Config) error {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	goalPct := cfg.GoalPercentage

	prompt := buildOneshotPrompt(feat.Slug, specPath, planPath, tasksPath, constitutionPath, projectRoot, brainstormText, goalPct)

	if oneshotCopy {
		if err := copyToClipboard(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✓ Copied agent prompt to clipboard")
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Paste the prompt to your coding agent\n")
		fmt.Printf("  2. The agent will drive clarification and fill out all documents\n")
		fmt.Printf("  3. Review the completed SPEC.md, PLAN.md, and TASKS.md\n")
		fmt.Printf("  4. Run 'kit implement %s' to begin execution\n", feat.Slug)
		return nil
	}

	fmt.Println()
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "🚀 Oneshot: All artifacts created, combined prompt ready" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()
	fmt.Println(whiteBold + "Created artifacts:" + reset)
	fmt.Printf("  • SPEC:  %s\n", specPath)
	fmt.Printf("  • PLAN:  %s\n", planPath)
	fmt.Printf("  • TASKS: %s\n", tasksPath)
	fmt.Println()
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "✅ Copy this prompt to your coding agent:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Print(prompt)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Copy the prompt above and paste it to your coding agent\n")
	fmt.Printf("  2. The agent will drive clarification and fill out all documents\n")
	fmt.Printf("  3. Review the completed SPEC.md, PLAN.md, and TASKS.md\n")
	fmt.Printf("  4. Run 'kit implement %s' to begin execution\n", feat.Slug)

	return nil
}

// buildOneshotPrompt assembles the full 5-phase agent prompt text.
func buildOneshotPrompt(slug, specPath, planPath, tasksPath, constitutionPath, projectRoot, brainstormText string, goalPct int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`# Oneshot: %s

You are driving the entire spec-driven development workflow for feature: **%s**

All artifact files have been created and are empty templates. Your job is to
read the brainstorming specification below, ask clarifying questions, and
progressively fill out each document until you reach >= %d%% understanding
of both the problem AND the solution.

## Document Hierarchy

| Document | Purpose | Focus |
|----------|---------|-------|
| CONSTITUTION.md | Project-wide constraints, principles, priors | Invariants |
| SPEC.md | Requirements, goals, acceptance criteria | WHAT to build |
| PLAN.md | Architecture, components, design decisions | HOW to build it |
| TASKS.md | Ordered execution steps with acceptance criteria | WHAT to do next |

## Context Documents (read first)
- **CONSTITUTION**: %s
- **SPEC**: %s (empty template — defines WHAT)
- **PLAN**: %s (empty template — defines HOW)
- **TASKS**: %s (empty template — defines execution order)
- **Project root**: %s

`, slug, slug, goalPct, constitutionPath, specPath, planPath, tasksPath, projectRoot))

	// phase 1: understand & clarify
	sb.WriteString(fmt.Sprintf(`## Phase 1: Understand & Clarify

1. Read CONSTITUTION.md to understand project constraints and principles
2. Read the SPEC.md, PLAN.md, and TASKS.md template files to understand the expected document structure and sections
3. Read the **Brainstorming Specification** at the bottom of this prompt
4. Analyze the codebase at %s to understand existing patterns and architecture
5. Ask clarifying questions in batches of 10 until you reach >= %d%% understanding of:
   - The **problem** being solved
   - The **solution** approach
   - The **constraints** and edge cases

Question format requirements:
- Label each question as **Fact-finding** (inputs, outputs, constraints, invariants) or **Decision-required** (tradeoffs the user must choose)
- When appropriate, include your **preferred solution** as one option, clearly labeled with reasoning (performance, simplicity, safety, cost)
- Present viable alternatives alongside your recommendation
- Do not assume acceptance of your preferred solution without user confirmation

6. After each batch, state your current understanding percentage
7. Begin drafting SPEC.md as your understanding grows — save progress to the file after each clarification round
8. Continue until understanding >= %d%%

`, projectRoot, goalPct, goalPct))

	// phase 2: spec
	sb.WriteString(fmt.Sprintf(`## Phase 2: Write SPEC.md

Once understanding >= %d%%, finalize %s with all sections complete:

- **SUMMARY**: 1-2 sentences — information-dense, includes core problem, solution approach, and key constraint
  - Example: "Adds CSV export for user data with streaming support for large datasets (>100k rows). Must complete in <5s and handle Unicode."
- **PROBLEM**: What problem does this feature solve?
- **GOALS**: Measurable outcomes
- **NON-GOALS**: Explicitly out of scope
- **USERS**: Who will use this feature?
- **REQUIREMENTS**: What must be true for this to be complete?
- **ACCEPTANCE**: How do we verify it works?
- **EDGE-CASES**: Unusual scenarios to handle
- **OPEN-QUESTIONS**: Remaining unknowns

Rules:
- Keep language precise
- Focus on WHAT, not HOW
- Respect constraints from CONSTITUTION.md
- All content MUST be written to %s — do not leave specification content only in chat

After completing SPEC.md, present a brief summary and confirm with the user before proceeding to PLAN.md.

`, goalPct, specPath, specPath))

	// phase 3: plan
	sb.WriteString(fmt.Sprintf(`## Phase 3: Write PLAN.md

After SPEC.md is approved, write %s:

- **SUMMARY**: One-paragraph overview of the approach
- **APPROACH**: Strategy, tradeoff decisions, no code
- **COMPONENTS**: Logical modules with clear responsibility boundaries
- **DATA**: Data shapes, structures, storage decisions
- **INTERFACES**: Commands, inputs, outputs, side effects
- **RISKS**: Technical risks with mitigation strategies
- **TESTING**: Validation strategy and test types

Rules:
- Focus on HOW, not WHAT (SPEC covers WHAT)
- Do not restate requirements
- No new scope beyond SPEC.md
- Avoid code unless strictly necessary
- PLAN.md must make TASKS.md obvious and deterministic
- All content MUST be written to %s — do not leave plan content only in chat

`, planPath, planPath))

	// phase 4: tasks
	sb.WriteString(fmt.Sprintf(`## Phase 4: Write TASKS.md

After PLAN.md is complete, write %s:

### A) Progress Table
| ID | TASK | STATUS | OWNER | DEPENDENCIES |

STATUS values: todo | doing | blocked | done

### B) Task List (markdown checkboxes)
- [ ] T001: task description
- [ ] T002: task description

IMPORTANT: Use exactly this checkbox format ('- [ ]' incomplete, '- [x]' complete).
'kit status' parses these checkboxes to track progress automatically.

### C) Task Details (per task)
### T00X
- **GOAL**: One sentence outcome
- **SCOPE**: Tight bullets
- **ACCEPTANCE**: Concrete checks
- **NOTES**: Only if necessary

### D) Dependencies
- Cross-task or external blockers
- Include the exact missing decision if applicable

### E) Notes
- Additional context or implementation notes (only if required to prevent ambiguity)

Rules:
- Tasks must be atomic and ordered
- Tasks must map back to PLAN items
- Tasks must imply an unambiguous implementation order
- A coding agent should execute them linearly with minimal back-and-forth
- All content MUST be written to %s — do not leave task content only in chat

`, tasksPath, tasksPath))

	// phase 5: pre-implementation review
	sb.WriteString(fmt.Sprintf(`## Phase 5: Pre-Implementation Review

After all documents are filled:

1. Review SPEC, PLAN, and TASKS for internal consistency
2. Verify SPEC → PLAN → TASKS traceability (every task traces to a plan item, every plan item traces to a spec requirement)
3. Confirm no scope creep beyond the brainstorming specification
4. State your final confidence level for each document
5. Present a brief summary of what will be implemented
6. **Ask the user for approval** before considering the documents finalized

The feature is now in **pre-implementation phase**, ready for execution
via 'kit implement %s' once documents are approved.

## IMPORTANT: File Update Requirement

All specification, plan, and task content MUST be written to the actual files:
- SPEC: %s
- PLAN: %s
- TASKS: %s

These files are the single source of truth for this feature.
Do not leave content only in chat — persist everything to the files.

## Workflow Rules

- Drive the entire process yourself — do not wait for step-by-step instructions
- Save progress to each file as you complete each phase
- If your understanding changes, update earlier documents before continuing
- If blocked on a question, explain what is blocking and suggest a resolution
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact
- Ensure all documents respect CONSTITUTION.md constraints

`, slug, specPath, planPath, tasksPath))

	// brainstorming specification section
	sb.WriteString(`---

## Brainstorming Specification

The following is the raw brainstorming specification provided by the user.
This is your primary input for understanding the feature.

`)
	sb.WriteString(brainstormText)
	sb.WriteString("\n")

	return sb.String()
}

// package cli implements the Kit command-line interface.
package cli

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/spf13/cobra"
)

var handoffCopy bool
var handoffOutputOnly bool

var handoffCmd = &cobra.Command{
	Use:   "handoff [feature]",
	Short: "Output context for a fresh coding agent session",
	Long: `Output instructions and context for starting a new coding agent session
with minimal information loss.

Use this when switching between agents (Warp, Claude, Copilot, Codex) due to
token limits or rate limiting. The output provides:
  - Kit project structure explanation
  - How to read and use Kit documents
  - Current project/feature state
  - Immediate next steps

Without a feature argument, outputs project-level context.
With a feature argument, outputs feature-specific context.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHandoff,
}

func init() {
	handoffCmd.Flags().BoolVarP(&handoffCopy, "copy", "c", false, "copy output to clipboard")
	handoffCmd.Flags().BoolVar(&handoffOutputOnly, "output-only", false, "output text only, suppressing status messages")
	rootCmd.AddCommand(handoffCmd)
}

func runHandoff(cmd *cobra.Command, args []string) error {
	var output string
	var err error

	if len(args) == 1 {
		output, err = featureHandoff(args[0])
	} else {
		output, err = projectHandoff()
	}

	if err != nil {
		return err
	}

	printWorkflowInstructions("handoff (supporting step)", []string{
		"resume your active phase: spec -> plan -> tasks -> implement -> reflect",
	})

	if handoffCopy {
		copyCmd := exec.Command("pbcopy")
		copyCmd.Stdin = strings.NewReader(output)
		if err := copyCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✓ Copied to clipboard")
		return nil
	}

	fmt.Print(output)
	return nil
}

func projectHandoff() (string, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		// no kit project, output generic instructions
		return genericHandoffInstructions(), nil
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := cfg.SpecsPath(projectRoot)
	features, _ := feature.ListFeatures(specsDir)

	var sb strings.Builder

	sb.WriteString("# Agent Context Handoff\n\n")
	sb.WriteString("You are continuing work on a Kit-managed project. Kit is a spec-driven development framework.\n\n")

	sb.WriteString("## How Kit Works\n\n")
	sb.WriteString("Kit enforces a document-driven development workflow:\n")
	sb.WriteString("1. **Constitution** (`docs/CONSTITUTION.md`) — project-wide constraints, principles, priors\n")
	sb.WriteString("2. **Specification** (`SPEC.md`) — what is being built and why\n")
	sb.WriteString("3. **Plan** (`PLAN.md`) — how it will be built\n")
	sb.WriteString("4. **Tasks** (`TASKS.md`) — executable work units\n")
	sb.WriteString("5. **Implementation** — code that fulfills the spec\n")
	sb.WriteString("6. **Reflection** — verify correctness, loop back if needed\n\n")

	sb.WriteString("## Key Principle\n\n")
	sb.WriteString("**Specs drive code. Code serves specs.**\n\n")
	sb.WriteString("Before implementing anything:\n")
	sb.WriteString("1. Read the relevant `SPEC.md` → `PLAN.md` → `TASKS.md`\n")
	sb.WriteString("2. Implement tasks in order\n")
	sb.WriteString("3. If reality diverges from spec, update spec first, then code\n\n")

	sb.WriteString("## Project Structure\n\n")
	sb.WriteString("```\n")
	sb.WriteString(projectRoot + "/\n")
	sb.WriteString("├── .kit.yaml              # Kit configuration\n")
	sb.WriteString("├── docs/\n")
	sb.WriteString("│   ├── CONSTITUTION.md    # Project constraints and principles\n")
	sb.WriteString("│   ├── PROJECT_PROGRESS_SUMMARY.md\n")
	sb.WriteString("│   └── specs/             # Feature specifications\n")

	if len(features) > 0 {
		for _, f := range features {
			sb.WriteString(fmt.Sprintf("│       └── %s/\n", f.DirName))
		}
	} else {
		sb.WriteString("│       └── <feature-dirs>/\n")
	}
	sb.WriteString("```\n\n")

	// read constitution summary if exists
	constitutionPath := cfg.ConstitutionAbsPath(projectRoot)
	if document.Exists(constitutionPath) {
		sb.WriteString("## Constitution Summary\n\n")
		sb.WriteString(fmt.Sprintf("Read `%s` for project-wide constraints.\n\n", cfg.ConstitutionPath))
	}

	// list features with their phases
	if len(features) > 0 {
		sb.WriteString("## Features\n\n")
		sb.WriteString("| Feature | Phase | Path |\n")
		sb.WriteString("| ------- | ----- | ---- |\n")
		for _, f := range features {
			sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", f.Slug, f.Phase, f.Path))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Immediate Actions\n\n")
	sb.WriteString("1. Read `docs/CONSTITUTION.md` to understand project constraints\n")
	if len(features) > 0 {
		sb.WriteString("2. Read `docs/PROJECT_PROGRESS_SUMMARY.md` for current state\n")
		sb.WriteString("3. Pick a feature and read its `SPEC.md` → `PLAN.md` → `TASKS.md`\n")
	} else {
		sb.WriteString("2. Run `kit spec <feature-name>` to create your first feature\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Kit Commands\n\n")
	sb.WriteString("- `kit spec <feature>` — create/open specification\n")
	sb.WriteString("- `kit plan <feature>` — create/open implementation plan\n")
	sb.WriteString("- `kit tasks <feature>` — create/open task list\n")
	sb.WriteString("- `kit check <feature>` — validate documents\n")
	sb.WriteString("- `kit summarize` — get context summarization instructions\n")
	sb.WriteString("- `kit reflect` — get reflection/verification instructions\n")

	return sb.String(), nil
}

func featureHandoff(featureRef string) (string, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return "", err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := cfg.SpecsPath(projectRoot)
	feat, err := feature.Resolve(specsDir, featureRef)
	if err != nil {
		return "", fmt.Errorf("feature '%s' not found: %w", featureRef, err)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Agent Context Handoff — Feature: %s\n\n", feat.Slug))
	sb.WriteString("You are continuing work on a specific feature in a Kit-managed project.\n\n")

	sb.WriteString("## Kit Workflow Reminder\n\n")
	sb.WriteString("**Specs drive code. Code serves specs.**\n\n")
	sb.WriteString("1. Read `docs/CONSTITUTION.md` for project constraints\n")
	sb.WriteString("2. Read `SPEC.md` → `PLAN.md` → `TASKS.md` in order\n")
	sb.WriteString("3. Implement tasks as defined\n")
	sb.WriteString("4. If reality diverges, update specs first, then code\n\n")

	sb.WriteString("## Feature Location\n\n")
	sb.WriteString(fmt.Sprintf("- **Path**: `%s`\n", feat.Path))
	sb.WriteString(fmt.Sprintf("- **Phase**: %s\n", feat.Phase))
	sb.WriteString(fmt.Sprintf("- **Directory**: %s\n\n", feat.DirName))

	sb.WriteString("## Required Reading\n\n")
	sb.WriteString("Read these documents in order:\n\n")

	// check which documents exist and provide guidance
	constitutionPath := cfg.ConstitutionAbsPath(projectRoot)
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	analysisPath := filepath.Join(feat.Path, "ANALYSIS.md")

	// always list CONSTITUTION first
	sb.WriteString(fmt.Sprintf("0. **CONSTITUTION.md** — `%s`\n", constitutionPath))
	sb.WriteString("   - Project-wide constraints and principles\n")
	sb.WriteString("   - **Read this first to understand fundamental rules**\n\n")

	if document.Exists(specPath) {
		sb.WriteString(fmt.Sprintf("1. **SPEC.md** — `%s`\n", specPath))
		sb.WriteString("   - Requirements and acceptance criteria\n")
		sb.WriteString("   - What problem we're solving\n")
		sb.WriteString("   - Edge cases and constraints\n\n")
	}

	if document.Exists(planPath) {
		sb.WriteString(fmt.Sprintf("2. **PLAN.md** — `%s`\n", planPath))
		sb.WriteString("   - Implementation approach\n")
		sb.WriteString("   - Component design\n")
		sb.WriteString("   - Technical decisions\n\n")
	}

	if document.Exists(tasksPath) {
		sb.WriteString(fmt.Sprintf("3. **TASKS.md** — `%s`\n", tasksPath))
		sb.WriteString("   - Work units and their status\n")
		sb.WriteString("   - Dependencies between tasks\n")
		sb.WriteString("   - **Start here for what to do next**\n\n")
	}

	if document.Exists(analysisPath) {
		sb.WriteString(fmt.Sprintf("4. **ANALYSIS.md** — `%s`\n", analysisPath))
		sb.WriteString("   - Understanding percentage\n")
		sb.WriteString("   - Open questions and assumptions\n\n")
	}

	sb.WriteString("## Immediate Actions\n\n")

	switch feat.Phase {
	case feature.PhaseSpec:
		sb.WriteString("1. Read SPEC.md thoroughly\n")
		sb.WriteString("2. Ask clarifying questions until understanding >= 95%\n")
		sb.WriteString("3. When ready, run `kit plan " + feat.Slug + "`\n")
	case feature.PhasePlan:
		sb.WriteString("1. Read SPEC.md and PLAN.md\n")
		sb.WriteString("2. Verify plan aligns with spec requirements\n")
		sb.WriteString("3. When ready, run `kit tasks " + feat.Slug + "`\n")
	case feature.PhaseTasks:
		sb.WriteString("1. Read TASKS.md to find incomplete tasks\n")
		sb.WriteString("2. Implement tasks in dependency order\n")
		sb.WriteString("3. Run `kit reflect " + feat.Slug + "` after implementation\n")
	default:
		sb.WriteString("1. Read all feature documents\n")
		sb.WriteString("2. Check TASKS.md for current status\n")
		sb.WriteString("3. Continue implementation or run `kit check " + feat.Slug + "`\n")
	}

	sb.WriteString("\n## Context Commands\n\n")
	sb.WriteString(fmt.Sprintf("- `kit summarize %s` — get summarization instructions\n", feat.Slug))
	sb.WriteString(fmt.Sprintf("- `kit reflect %s` — get reflection/verification instructions\n", feat.Slug))
	sb.WriteString(fmt.Sprintf("- `kit check %s` — validate feature documents\n", feat.Slug))

	return sb.String(), nil
}

func genericHandoffInstructions() string {
	return `# Agent Context Handoff

You are starting work on a project that may use Kit for spec-driven development.

## What is Kit?

Kit is a document-centered CLI that enforces a specification-driven workflow:

1. **Constitution** — project-wide constraints, principles, priors
2. **Specification** — what is being built and why
3. **Plan** — how it will be built
4. **Tasks** — executable work units
5. **Implementation** — code execution
6. **Reflection** — verify correctness, refine understanding

## Key Principle

**Specs drive code. Code serves specs.**

Before implementing:
1. Read SPEC.md → PLAN.md → TASKS.md
2. Implement tasks in order
3. If reality diverges, update spec first, then code

## Check for Kit Project

Run these commands to check if this is a Kit project:
` + "```bash" + `
# check for Kit config
ls -la .kit.yaml

# check for Kit documents
ls -la docs/CONSTITUTION.md
ls -la docs/specs/

# if Kit project exists, get full context
kit handoff
` + "```" + `

## If Not a Kit Project

Initialize with:
` + "```bash" + `
kit init
kit spec <first-feature>
` + "```" + `

## Kit Commands

- ` + "`kit init`" + ` — initialize project
- ` + "`kit spec <feature>`" + ` — create specification
- ` + "`kit plan <feature>`" + ` — create implementation plan
- ` + "`kit tasks <feature>`" + ` — create task list
- ` + "`kit check <feature>`" + ` — validate documents
- ` + "`kit handoff [feature]`" + ` — output this context (use when switching agents)
`
}

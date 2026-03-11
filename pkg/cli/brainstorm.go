package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var (
	brainstormCopy       bool
	brainstormOutput     string
	brainstormOutputOnly bool
)

var brainstormCmd = &cobra.Command{
	Use:   "brainstorm [feature]",
	Short: "Create or update BRAINSTORM.md and output a planning prompt",
	Long: `Create or update a feature's BRAINSTORM.md document and output a
planning-only prompt for a coding agent.

Creates:
	- Feature directory (e.g., docs/specs/0001-my-feature/)
	- BRAINSTORM.md as the first feature-scoped artifact

Interactive flow:
	1. Ask for a feature/project name (unless provided as an argument)
	2. Ask for a short issue/feature thesis

The command never implements code. It outputs a /plan prompt that instructs
the coding agent to research the codebase, use numbered lists for clarifying
questions, show percentage progress, and persist findings to BRAINSTORM.md.

Examples:
	kit brainstorm
	kit brainstorm patient-intake-redesign --copy
	kit brainstorm -o docs/brainstorm-prompt.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBrainstorm,
}

func init() {
	brainstormCmd.Flags().BoolVar(&brainstormCopy, "copy", false, "copy output to clipboard")
	brainstormCmd.Flags().StringVarP(&brainstormOutput, "output", "o", "", "write output to file")
	brainstormCmd.Flags().BoolVar(&brainstormOutputOnly, "output-only", false, "output text only, suppressing status messages")
	rootCmd.AddCommand(brainstormCmd)
}

func runBrainstorm(cmd *cobra.Command, args []string) error {
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
	if err := ensureDir(specsDir); err != nil {
		return err
	}

	featureRef, thesis, err := promptBrainstormInputs(args)
	if err != nil {
		return err
	}

	feat, created, err := feature.EnsureExists(cfg, specsDir, featureRef)
	if err != nil {
		return err
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if !document.Exists(brainstormPath) {
		if err := document.Write(brainstormPath, templates.BuildBrainstormArtifact(thesis)); err != nil {
			return fmt.Errorf("failed to create BRAINSTORM.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created BRAINSTORM.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ BRAINSTORM.md already exists")
	}

	if !outputOnly {
		if created {
			fmt.Printf("📁 Created feature directory: %s\n", feat.DirName)
		} else {
			fmt.Printf("📁 Using existing feature: %s\n", feat.DirName)
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		if !outputOnly {
			fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		}
	} else if !outputOnly {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	prompt := buildBrainstormPrompt(brainstormPath, feat.Slug, projectRoot, thesis, cfg.GoalPercentage)

	if brainstormOutput != "" {
		if err := document.Write(brainstormOutput, prompt); err != nil {
			return fmt.Errorf("failed to write prompt file: %w", err)
		}
		if !outputOnly {
			fmt.Printf("✓ Written prompt to %s\n", brainstormOutput)
		}
	}

	if brainstormOutput == "" {
		if err := outputPrompt(prompt, outputOnly, brainstormCopy); err != nil {
			return err
		}
	} else if brainstormCopy {
		if err := copyToClipboard(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		if !outputOnly {
			fmt.Println("Copied agent instructions to clipboard.")
		}
	}

	if !outputOnly {
		printWorkflowInstructions("brainstorm (optional pre-spec)", []string{
			fmt.Sprintf("review and refine %s", brainstormPath),
			fmt.Sprintf("run kit spec %s when the brainstorm is complete", feat.Slug),
			"then continue spec -> plan -> tasks -> implement -> reflect",
		})
	}

	return nil
}

func promptBrainstormInputs(args []string) (string, string, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:              whiteBold + "   > " + reset,
		InterruptPrompt:     "^C",
		EOFPrompt:           "",
		Stdin:               os.Stdin,
		Stdout:              os.Stdout,
		Stderr:              os.Stderr,
		FuncFilterInputRune: specInputRuneFilter,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	featureRef := ""
	if len(args) == 1 {
		featureRef = normalizeSpecAnswer(args[0])
	}

	if featureRef == "" {
		fmt.Println()
		fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
		fmt.Println(whiteBold + "🧠 Brainstorm Builder" + reset)
		fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
		fmt.Println()
		fmt.Println(dim + "Step 1 of 2: Enter a feature/project name." + reset)
		fmt.Println(dim + "It will be normalized to lowercase kebab-case and must be 5 words or fewer." + reset)
		featureRef = readLineRL(rl)
	}

	if featureRef == "" {
		return "", "", fmt.Errorf("feature name cannot be empty")
	}

	normalized := feature.NormalizeSlug(featureRef)
	if err := feature.ValidateSlug(normalized); err != nil {
		return "", "", err
	}

	if normalized != featureRef {
		fmt.Printf(dim+"Using normalized feature slug: %s"+reset+"\n\n", normalized)
	}

	fmt.Println(dim + "Step 2 of 2: Describe the issue or feature in a few sentences." + reset)
	fmt.Println(dim + "Press Enter to submit. Use Shift+Enter or Ctrl+J to insert a newline." + reset)
	thesis := readLineRL(rl)
	if thesis == "" {
		return "", "", fmt.Errorf("brainstorm thesis cannot be empty")
	}

	return normalized, thesis, nil
}

func buildBrainstormPrompt(brainstormPath, featureSlug, projectRoot, thesis string, goalPct int) string {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")

	var sb strings.Builder
	sb.WriteString("/plan\n\n")
	sb.WriteString(fmt.Sprintf(`You are in planning mode for feature: **%s**

You MUST update the brainstorm file at:
- **BRAINSTORM**: %s
- **Feature**: %s
- **Project Root**: %s

## User Thesis

%s

## Context Docs (read first)
| File | Purpose |
|------|---------|
| CONSTITUTION | %s |
| BRAINSTORM | %s |
| Project Root | %s |

## Your Task

1. Stay in planning and information-gathering mode only
2. Do NOT implement code, write production changes, or move into execution
3. Read CONSTITUTION.md first to understand project constraints and workflow rules
4. Read the current BRAINSTORM.md template and treat it as the source of truth for this research phase
5. Research the entire codebase at %s to identify relevant files, patterns, constraints, interfaces, and adjacent workflows
6. Ask clarifying questions until you reach ≥%d%% confidence that you understand the problem and desired solution
7. Use numbered lists
8. Ask questions in batches of up to 10
9. For every question, include your current best proposed solution or assumption
10. State uncertainties
11. After each batch of up to 10 questions, output your current percentage understanding so the user can see progress
12. Reassess, update %s with durable findings, and continue with additional batches of up to 10 questions until the specification is precise enough to produce a correct, production-quality solution
13. Keep every finding filepath-specific whenever possible
14. If you create a tentative plan in chat, fold the durable conclusions back into %s so the file stays current
15. Stop before implementation. The next workflow step after this research phase is usually kit spec %s

## BRAINSTORM.md Requirements

The final BRAINSTORM.md must be a detailed, informational, filepath-specific document with:
- SUMMARY
- USER THESIS
- CODEBASE FINDINGS
- AFFECTED FILES
- QUESTIONS
- OPTIONS
- RECOMMENDED STRATEGY
- NEXT STEP

## Rules

- planning only — no implementation
- no build or execution work intended to advance code changes
- the purpose of this phase is understanding, not code output
- use numbered lists for clarifying questions and progress updates
- continue the clarification loop until confidence reaches ≥%d%% and the specification is precise enough for a correct, production-quality solution
- preserve facts in BRAINSTORM.md, not just in chat
- make the final document dense, explicit, and easy for a coding agent to use when drafting SPEC.md
`, featureSlug, brainstormPath, featureSlug, projectRoot, thesis, constitutionPath, brainstormPath, projectRoot, projectRoot, goalPct, brainstormPath, brainstormPath, featureSlug, goalPct))

	return sb.String()
}

// copyToClipboard copies text to the system clipboard.
func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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

var (
	brainstormCopy       bool
	brainstormEditor     string
	brainstormOutput     string
	brainstormOutputOnly bool
	brainstormUseVim     bool
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
	2. Ask for a multiline issue/feature thesis

The command never implements code. It outputs a /plan prompt that instructs
the coding agent to research the codebase, use numbered lists for clarifying
questions, show percentage progress, and persist findings to BRAINSTORM.md.

Examples:
	kit brainstorm
	kit brainstorm --vim
	kit brainstorm patient-intake-redesign
	kit brainstorm patient-intake-redesign --output-only
	kit brainstorm -o docs/brainstorm-prompt.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBrainstorm,
}

func init() {
	addFreeTextInputFlags(brainstormCmd, &brainstormUseVim, &brainstormEditor)
	brainstormCmd.Flags().BoolVar(&brainstormCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	brainstormCmd.Flags().StringVarP(&brainstormOutput, "output", "o", "", "write output to file")
	brainstormCmd.Flags().BoolVar(&brainstormOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(brainstormCmd)
	rootCmd.AddCommand(brainstormCmd)
}

func runBrainstorm(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	promptOnly := promptOnlyEnabled(cmd)

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

	if promptOnly {
		if brainstormUseVim || brainstormEditor != "" {
			return fmt.Errorf("--prompt-only cannot be used with --vim or --editor")
		}
		return outputExistingBrainstormPrompt(args, projectRoot, cfg, outputOnly)
	}

	featureRef, thesis, err := promptBrainstormInputs(
		args,
		newFreeTextInputConfig(brainstormUseVim, brainstormEditor),
	)
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
	preparedPrompt := prepareAgentPrompt(prompt)

	if brainstormOutput != "" {
		if err := document.Write(brainstormOutput, preparedPrompt); err != nil {
			return fmt.Errorf("failed to write prompt file: %w", err)
		}
		if !outputOnly {
			fmt.Printf("✓ Written prompt to %s\n", brainstormOutput)
		}
	}

	if err := writePromptWithClipboardDefault(preparedPrompt, outputOnly, brainstormCopy); err != nil {
		return err
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

func outputExistingBrainstormPrompt(args []string, projectRoot string, cfg *config.Config, outputOnly bool) error {
	if brainstormOutput != "" {
		return fmt.Errorf("--prompt-only cannot be used with --output because it writes a file")
	}

	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 1 {
		feat, err = feature.Resolve(specsDir, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found: %w", args[0], err)
		}
	} else {
		feat, err = selectFeatureForBrainstormPromptOnly(specsDir)
		if err != nil {
			return err
		}
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if !document.Exists(brainstormPath) {
		return fmt.Errorf("BRAINSTORM.md not found for %s. Run 'kit brainstorm %s' first", feat.Slug, feat.Slug)
	}

	thesis := existingBrainstormThesis(brainstormPath)
	prompt := buildBrainstormPrompt(brainstormPath, feat.Slug, projectRoot, thesis, cfg.GoalPercentage)
	preparedPrompt := prepareAgentPrompt(prompt)

	if brainstormOutput != "" {
		if err := document.Write(brainstormOutput, preparedPrompt); err != nil {
			return fmt.Errorf("failed to write prompt file: %w", err)
		}
		if !outputOnly {
			fmt.Printf("✓ Written prompt to %s\n", brainstormOutput)
		}
	}

	if err := writePromptWithClipboardDefault(preparedPrompt, outputOnly, brainstormCopy); err != nil {
		return err
	}

	if !outputOnly {
		printWorkflowInstructions("brainstorm (existing feature prompt)", []string{
			fmt.Sprintf("review and refine %s", brainstormPath),
			fmt.Sprintf("run kit spec %s when the brainstorm is complete", feat.Slug),
			"no repository docs were mutated by this prompt-only run",
		})
	}

	return nil
}

func promptBrainstormInputs(args []string, inputCfg freeTextInputConfig) (string, string, error) {
	featureRef, err := promptBrainstormFeatureRef(args)
	if err != nil {
		return "", "", err
	}

	thesis, err := promptBrainstormThesis(inputCfg)
	if err != nil {
		return "", "", err
	}

	return featureRef, thesis, nil
}

func selectFeatureForBrainstormPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "BRAINSTORM.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features with BRAINSTORM.md available\n\nRun 'kit brainstorm <feature>' first")
	}

	fmt.Println()
	fmt.Println(whiteBold + "Select a feature to regenerate the brainstorm prompt for:" + reset)
	fmt.Println()
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
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

func existingBrainstormThesis(brainstormPath string) string {
	doc, err := document.ParseFile(brainstormPath, document.TypeBrainstorm)
	if err != nil {
		return "Continue the existing brainstorm using the current file contents as the source of truth."
	}

	if section := doc.GetSection("USER THESIS"); section != nil {
		if thesis := document.ExtractFirstParagraph(section); thesis != "" {
			return thesis
		}
	}
	if section := doc.GetSection("SUMMARY"); section != nil {
		if summary := document.ExtractFirstParagraph(section); summary != "" {
			return summary
		}
	}

	return "Continue the existing brainstorm using the current file contents as the source of truth."
}

func promptBrainstormFeatureRef(args []string) (string, error) {

	featureRef := ""
	if len(args) == 1 {
		featureRef = normalizeSpecAnswer(args[0])
	}
	if featureRef == "" {
		rl, err := newMultilineReadline()
		if err != nil {
			return "", fmt.Errorf("failed to initialize readline: %w", err)
		}
		defer closeMultilineReadline(rl)
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
		return "", fmt.Errorf("feature name cannot be empty")
	}

	normalized := feature.NormalizeSlug(featureRef)
	if err := feature.ValidateSlug(normalized); err != nil {
		return "", err
	}

	if normalized != featureRef {
		fmt.Printf(dim+"Using normalized feature slug: %s"+reset+"\n\n", normalized)
	}
	return normalized, nil
}

func promptBrainstormThesis(inputCfg freeTextInputConfig) (string, error) {

	fmt.Println(dim + "Step 2 of 2: Describe the issue or feature in a few sentences." + reset)
	if inputCfg.usesEditor() {
		fmt.Println(dim + "A vim-compatible editor will open for this response." + reset)
		return readEditorText(inputCfg, "brainstorm thesis", false)
	}

	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	fmt.Println(dim + "Press Enter to submit. Use Shift+Enter or Ctrl+J to insert newlines." + reset)
	fmt.Println(dim + "Consecutive blank lines are preserved." + reset)
	thesis := readLineRL(rl)
	if thesis == "" {
		return "", fmt.Errorf("brainstorm thesis cannot be empty")
	}

	return thesis, nil
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
`, featureSlug, brainstormPath, featureSlug, projectRoot, thesis, constitutionPath, brainstormPath, projectRoot, projectRoot))

	sb.WriteString(fmt.Sprintf("6. Keep the `## DEPENDENCIES` table in %s current throughout the research phase:\n", brainstormPath))
	sb.WriteString("   - include every dependency that materially shapes the feature definition, such as skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, and assets\n")
	sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
	sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
	sb.WriteString("   - for Figma or other MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
	sb.WriteString("   - if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale` instead of deleting it\n")

	nextStep := appendNumberedSteps(
		&sb,
		7,
		clarificationLoopSteps(
			goalPct,
			fmt.Sprintf(
				"Reassess, update %s with durable findings, and continue with "+
					"additional batches of up to 10 questions until the specification "+
					"is precise enough to produce a correct, production-quality solution",
				brainstormPath,
			),
		),
	)

	sb.WriteString(fmt.Sprintf(`%d. Keep every finding filepath-specific whenever possible
%d. If you create a tentative plan in chat, fold the durable conclusions back into %s so the file stays current
%d. Stop before implementation. The next workflow step after this research phase is usually kit spec %s

## BRAINSTORM.md Requirements

The final BRAINSTORM.md must be a detailed, informational, filepath-specific document with:
- SUMMARY
- USER THESIS
- CODEBASE FINDINGS
- AFFECTED FILES
- DEPENDENCIES
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
- keep the ## DEPENDENCIES table aligned with the tools, docs, and design references used during the phase
`, nextStep, nextStep+1, brainstormPath, nextStep+2, featureSlug, goalPct))
	appendNonEmptySectionRules(&sb, "`BRAINSTORM.md`")

	return sb.String()
}

// copyToClipboard copies text to the system clipboard.
func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

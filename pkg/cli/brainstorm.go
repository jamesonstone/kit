package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var (
	brainstormCopy       bool
	brainstormBacklog    bool
	brainstormInline     bool
	brainstormEditor     string
	brainstormOutput     string
	brainstormOutputOnly bool
	brainstormPickup     bool
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
	2. Open a vim-compatible editor for the multiline issue/feature thesis by default

The command never implements code. It outputs a /plan prompt that instructs
the coding agent to research the codebase, use numbered lists for clarifying
questions, show percentage progress, and persist findings to BRAINSTORM.md.

Examples:
	kit brainstorm
	kit brainstorm --inline
	kit brainstorm --editor nvim
	kit brainstorm patient-intake-redesign
	kit brainstorm patient-intake-redesign --output-only
	kit brainstorm -o docs/brainstorm-prompt.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBrainstorm,
}

func init() {
	addFreeTextInputFlags(brainstormCmd, &brainstormUseVim, &brainstormEditor)
	addInlineTextInputFlag(brainstormCmd, &brainstormInline)
	brainstormCmd.Flags().BoolVar(&brainstormBacklog, "backlog", false, "capture a deferred brainstorm item and leave it paused")
	brainstormCmd.Flags().BoolVar(&brainstormCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	brainstormCmd.Flags().StringVarP(&brainstormOutput, "output", "o", "", "write output to file")
	brainstormCmd.Flags().BoolVar(&brainstormOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	brainstormCmd.Flags().BoolVar(&brainstormPickup, "pickup", false, "resume a paused backlog item instead of starting a new brainstorm capture")
	if flag := brainstormCmd.Flags().Lookup("pickup"); flag != nil {
		flag.Hidden = true
		flag.Deprecated = "use `kit resume <feature>` or `kit backlog --pickup <feature>`"
	}
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

	if brainstormBacklog && brainstormPickup {
		return fmt.Errorf("--backlog and --pickup cannot be used together")
	}

	if promptOnly {
		if brainstormUseVim || brainstormEditor != "" || brainstormInline || brainstormBacklog || brainstormPickup {
			return fmt.Errorf("--prompt-only cannot be used with --vim, --editor, --inline, --backlog, or --pickup")
		}
		return outputExistingBrainstormPrompt(args, projectRoot, cfg, outputOnly)
	}

	if brainstormInline && (brainstormUseVim || brainstormEditor != "") {
		return fmt.Errorf("--inline cannot be used with --vim or --editor")
	}
	if brainstormPickup {
		if brainstormUseVim || brainstormEditor != "" || brainstormInline {
			return fmt.Errorf("--pickup cannot be used with --vim, --editor, or --inline")
		}
		return runBrainstormPickup(projectRoot, cfg, specsDir, args, outputOnly)
	}
	if brainstormBacklog {
		if brainstormOutput != "" || outputOnly || brainstormCopy {
			return fmt.Errorf("--backlog capture does not output a prompt; remove --output, --output-only, and --copy")
		}
		return runBrainstormBacklog(projectRoot, cfg, specsDir, args)
	}

	featureRef, thesis, err := promptBrainstormInputs(
		args,
		newFreeTextInputConfig(brainstormUseVim, brainstormEditor, brainstormInline, true),
	)
	if err != nil {
		return err
	}

	feat, created, err := feature.EnsureExists(cfg, specsDir, featureRef)
	if err != nil {
		return err
	}
	feature.ApplyLifecycleState(feat, cfg)

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

	wasPaused := feat.Paused
	if !created {
		if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
			return err
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
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
		printWorkflowInstructions("brainstorm (optional pre-spec)", []string{
			fmt.Sprintf("review and refine %s", brainstormPath),
			fmt.Sprintf("run kit spec %s when the brainstorm is complete", feat.Slug),
			"then continue spec -> plan -> tasks -> implement -> reflect",
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
		style := styleForStdout()
		printSectionBanner("🧠", "Brainstorm Builder")
		fmt.Println(style.muted("Step 1 of 2: Enter a feature/project name."))
		fmt.Println(style.muted("It will be normalized to lowercase kebab-case and must be 5 words or fewer."))
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
	style := styleForStdout()

	fmt.Println()
	fmt.Println(style.muted("Step 2 of 2: Describe the issue or feature in a few sentences."))
	if inputCfg.usesEditor() {
		fmt.Printf("%s\n", style.muted(fmt.Sprintf("A %s will open for this response.", inputCfg.editorLabel())))
		return readEditorText(inputCfg, "brainstorm thesis", false)
	}

	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	fmt.Println(style.muted("Press Enter to submit. Use Shift+Enter or Ctrl+J to insert newlines."))
	fmt.Println(style.muted("Consecutive blank lines are preserved."))
	thesis := readLineRL(rl)
	if thesis == "" {
		return "", fmt.Errorf("brainstorm thesis cannot be empty")
	}

	return thesis, nil
}

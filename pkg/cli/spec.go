package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var specCopy bool
var specEditor string
var specInline bool
var specOutputOnly bool
var specUseVim bool

var specCmd = &cobra.Command{
	Use:   "spec [feature]",
	Short: "Create or open a feature specification",
	Long: `Create a new feature specification or open an existing one.

Creates:
  - Feature directory (e.g., docs/specs/0001-my-feature/)
  - SPEC.md with required sections and placeholder comments

Updates PROJECT_PROGRESS_SUMMARY.md after creation.

If no feature is specified, shows an interactive selection of existing
features with BRAINSTORM.md or SPEC.md.

Modes:
  Default:        Copy the generated prompt to the clipboard and show status (non-interactive)
  --interactive:  Prompt user for spec details, opening a vim-compatible editor for free-text answers by default
  --template:     Output empty template without interactive questions (deprecated, same as default)

Flags:
  --output-only:  Output the raw prompt to stdout instead of copying it to the clipboard
  --copy:         Copy prompt to clipboard (mainly useful with --output-only)
  --interactive:  Force interactive prompts even when stdin is not a terminal
  --vim:          Open free-text responses in a vim-compatible editor
  --inline:       Use inline multiline prompts instead of opening the editor`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSpec,
}

func init() {
	addFreeTextInputFlags(specCmd, &specUseVim, &specEditor)
	addInlineTextInputFlag(specCmd, &specInline)
	specCmd.Flags().Bool("template", false, "(deprecated) output empty template and prompt without interactive questions")
	specCmd.Flags().Bool("interactive", false, "prompt user for spec details interactively")
	specCmd.Flags().BoolVar(&specCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	specCmd.Flags().BoolVar(&specOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(specCmd)
	rootCmd.AddCommand(specCmd)
}

func runSpec(cmd *cobra.Command, args []string) error {
	specTemplateOnly, _ := cmd.Flags().GetBool("template")
	specInteractive, _ := cmd.Flags().GetBool("interactive")
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

	// ensure specs directory exists
	if err := ensureDir(specsDir); err != nil {
		return err
	}

	if promptOnly {
		if specTemplateOnly || specInteractive || specUseVim || specEditor != "" || specInline {
			return fmt.Errorf("--prompt-only cannot be used with --template, --interactive, --vim, --editor, or --inline")
		}
		return runSpecPromptOnly(args, projectRoot, cfg, outputOnly)
	}

	var (
		feat    *feature.Feature
		created bool
	)

	if len(args) == 0 {
		feat, err = selectFeatureForSpec(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {
		featureRef := args[0]

		// create or find feature
		feat, created, err = feature.EnsureExists(cfg, projectRoot, specsDir, featureRef)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	}

	if !outputOnly {
		if created {
			fmt.Printf("📁 Created feature directory: %s\n", feat.DirName)
		} else {
			fmt.Printf("📁 Using existing feature: %s\n", feat.DirName)
		}
	}

	// create SPEC.md if it doesn't exist
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		if err := document.Write(specPath, templates.Spec); err != nil {
			return fmt.Errorf("failed to create SPEC.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created SPEC.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ SPEC.md already exists")
	}

	// determine if we should run interactive mode
	// default is non-interactive (template mode), unless --interactive is explicitly set
	isInteractive := specInteractive && !specTemplateOnly
	inputCfg := newFreeTextInputConfig(specUseVim, specEditor, specInline, isInteractive)
	if (specUseVim || specEditor != "" || specInline) && !isInteractive {
		return fmt.Errorf("--vim, --editor, and --inline require --interactive")
	}
	if specInline && (specUseVim || specEditor != "") {
		return fmt.Errorf("--inline cannot be used with --vim or --editor")
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if !outputOnly && document.Exists(brainstormPath) {
		fmt.Println("  ✓ Found BRAINSTORM.md")
	}

	wasPaused := feat.Paused
	if !created {
		if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
			return err
		}
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
		fmt.Printf("\n✅ Feature '%s' ready!\n", feat.Slug)
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
	}

	if isInteractive {
		// interactive mode: gather details and compile prompt
		return runSpecInteractive(specPath, brainstormPath, feat, projectRoot, cfg, inputCfg, outputOnly)
	}

	// template mode: output the template and instructions
	return runSpecTemplate(specPath, brainstormPath, feat.Slug, projectRoot, cfg, outputOnly)
}

func runSpecPromptOnly(args []string, projectRoot string, cfg *config.Config, outputOnly bool) error {
	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 0 {
		feat, err = selectFeatureForSpecPromptOnly(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found: %w", args[0], err)
		}
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	return runSpecTemplate(specPath, brainstormPath, feat.Slug, projectRoot, cfg, outputOnly)
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

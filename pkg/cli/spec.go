package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

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
var specReviseThesis bool
var specUseVim bool

var promptSpecFeatureRef = readSpecFeatureRef

var specCmd = &cobra.Command{
	Use:   "spec [feature]",
	Short: "Run the v2 single-SPEC feature workflow",
	Long: `Create or open the v2 single-SPEC feature workflow.

Creates:
  - Feature directory (e.g., docs/specs/0001-my-feature/)
  - SPEC.md as the single durable feature workflow artifact
  - Feature notes/reference-material directories for supporting inputs

Updates PROJECT_PROGRESS_SUMMARY.md after creation.

If no feature is specified, shows an interactive selection of eligible existing
feature directories that can enter the v2 SPEC.md workflow. If there are no
eligible existing features to select from, prompts for a new feature name and
opens one thesis/goal editor entry before rendering the v2 supervisor prompt.

The generated prompt is the v2 supervisor contract. It keeps ideation,
clarification, implementation planning, task tracking, implementation,
validation, reflection, documentation updates, and delivery gating inside
SPEC.md instead of requiring separate BRAINSTORM.md, PLAN.md, TASKS.md,
implement, reflect, or standalone verification workflow commands.

Modes:
  New SPEC.md:       Ask for one required thesis/goal entry, capture delivery intent, then copy the v2 supervisor prompt
  Existing SPEC.md:  Preserve current SPEC.md content and regenerate/copy the v2 supervisor prompt
  --revise-thesis:   Append a dated thesis note to an existing SPEC.md before prompt output

Flags:
  --output-only:     Output the raw prompt to stdout instead of copying it to the clipboard
  --copy:            Copy prompt to clipboard (mainly useful with --output-only)
  --revise-thesis:   Append a dated thesis note and refresh delivery intent before prompt output
  --vim:             Open the thesis response in a vim-compatible editor instead of $EDITOR
  --inline:          Use inline multiline thesis entry instead of opening the editor`,
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
	specCmd.Flags().BoolVar(&specReviseThesis, "revise-thesis", false, "append a dated thesis note and refresh delivery intent before prompt output")
	addPromptOnlyFlag(specCmd)
	_ = specCmd.Flags().MarkHidden("template")
	_ = specCmd.Flags().MarkHidden("interactive")
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
		if specTemplateOnly || specInteractive || specReviseThesis || specUseVim || specEditor != "" || specInline {
			return fmt.Errorf("--prompt-only cannot be used with --template, --interactive, --revise-thesis, --vim, --editor, or --inline")
		}
		return runSpecPromptOnly(args, projectRoot, cfg, outputOnly)
	}
	if specTemplateOnly && specReviseThesis {
		return fmt.Errorf("--template cannot be used with --revise-thesis")
	}
	if specTemplateOnly && specInteractive {
		return fmt.Errorf("--template cannot be used with --interactive")
	}

	var (
		feat    *feature.Feature
		created bool
	)

	if len(args) == 0 {
		feat, err = selectFeatureForSpec(specsDir)
		if errors.Is(err, errNoSpecSelectionCandidates) {
			var featureRef string
			featureRef, err = promptSpecFeatureRef()
			if err != nil {
				return err
			}
			feat, created, err = feature.EnsureExists(cfg, projectRoot, specsDir, featureRef)
			if err != nil {
				return err
			}
			specInteractive = !specTemplateOnly
		} else if err != nil {
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

	specPath := filepath.Join(feat.Path, "SPEC.md")
	specWasCreated := false
	if !document.Exists(specPath) {
		content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(specPath, content); err != nil {
			return fmt.Errorf("failed to create SPEC.md: %w", err)
		}
		specWasCreated = true
		if !outputOnly {
			fmt.Println("  ✓ Created SPEC.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ SPEC.md already exists")
	}
	if changed, err := ensureSpecV2Adoption(specPath, projectRoot, feat.DirName); err != nil {
		return err
	} else if changed && !outputOnly {
		fmt.Println("  ✓ Updated SPEC.md for v2 workflow")
	}
	if effectivePromptProfile(feat.Path) == promptProfileFrontend {
		if _, _, err := ensureFeatureDesignMaterialsDirs(projectRoot, feat.DirName); err != nil {
			return err
		}
		if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
			return err
		}
	}

	needsThesisInput := !specTemplateOnly && (specWasCreated || specReviseThesis)
	inputCfg := newFreeTextInputConfig(specUseVim, specEditor, specInline, needsThesisInput)
	if specInline && (specUseVim || specEditor != "") {
		return fmt.Errorf("--inline cannot be used with --vim or --editor")
	}
	if specInteractive && !needsThesisInput {
		return fmt.Errorf("--interactive has been replaced by the default thesis prompt for new SPEC.md files; use --revise-thesis for existing SPEC.md files")
	}
	if (specUseVim || specEditor != "" || specInline) && !needsThesisInput {
		return fmt.Errorf("--vim, --editor, and --inline require a new SPEC.md or --revise-thesis")
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

	var compiledAnswers *specAnswers
	if needsThesisInput {
		compiledAnswers, err = runSpecInteractive(specPath, brainstormPath, feat, projectRoot, cfg, inputCfg, specWasCreated, outputOnly)
		if err != nil {
			return err
		}
		if !outputOnly {
			fmt.Println("  ✓ Captured thesis and delivery intent in SPEC.md")
		}
	} else if specReviseThesis && !outputOnly {
		fmt.Println("  ✓ Thesis revision skipped")
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

	if needsThesisInput {
		return outputCompiledPrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg, compiledAnswers, outputOnly)
	}

	// template mode: output the template and instructions
	return runSpecTemplate(specPath, brainstormPath, feat.Slug, projectRoot, cfg, outputOnly, false)
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
	return runSpecTemplate(specPath, brainstormPath, feat.Slug, projectRoot, cfg, outputOnly, true)
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func ensureSpecV2Adoption(specPath, projectRoot, featureDirName string) (bool, error) {
	before, err := os.ReadFile(specPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", specPath, err)
	}

	if err := document.MergeDocument(specPath, templates.Spec, document.TypeSpec); err != nil {
		return false, fmt.Errorf("failed to add v2 SPEC.md sections to %s: %w", specPath, err)
	}

	afterMerge, err := os.ReadFile(specPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s after v2 section adoption: %w", specPath, err)
	}

	_, notesRelPath, err := ensureFeatureNotesDir(projectRoot, featureDirName)
	if err != nil {
		return false, err
	}

	phase := "clarify"
	doc := document.Parse(string(afterMerge), specPath, document.TypeSpec)
	if doc.Metadata != nil && doc.Metadata.WorkflowVersion == 2 && doc.Metadata.Phase != "" {
		phase = doc.Metadata.Phase
	}

	updated, _, err := document.UpsertMetadata(string(afterMerge), document.TypeSpec, document.MetadataUpsert{
		Feature:         document.FeatureMetadataFromDir(featureDirName),
		WorkflowVersion: 2,
		Phase:           phase,
		References:      referencesForMetadataUpsert(string(afterMerge), document.TypeSpec, []document.MetadataReference{featureNotesReference(notesRelPath)}),
	})
	if err != nil {
		return false, fmt.Errorf("failed to update v2 SPEC.md metadata in %s: %w", specPath, err)
	}
	if updated != string(afterMerge) {
		if err := document.Write(specPath, updated); err != nil {
			return false, fmt.Errorf("failed to write v2 SPEC.md metadata in %s: %w", specPath, err)
		}
	}

	return string(before) != updated, nil
}

func readSpecFeatureRef() (string, error) {
	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	style := styleForStdout()
	fmt.Println(style.muted("No eligible v2 feature candidates found. Enter a feature/project name."))
	fmt.Println(style.muted("It will be normalized to lowercase kebab-case and must be 5 words or fewer."))
	featureRef := readLineRL(rl)
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

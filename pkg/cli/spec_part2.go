package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

func runSpec(cmd *cobra.Command, args []string) error {
	specTemplateOnly, _ := cmd.Flags().GetBool("template")
	specInteractive, _ := cmd.Flags().GetBool("interactive")
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	promptOnly := promptOnlyEnabled(cmd)

	if promptOnly {
		if specTemplateOnly || specInteractive || specReviseThesis || specUseVim || specEditor != "" || specInline {
			return fmt.Errorf("--prompt-only cannot be used with --template, --interactive, --revise-thesis, --vim, --editor, or --inline")
		}
	}
	if specTemplateOnly && specReviseThesis {
		return fmt.Errorf("--template cannot be used with --revise-thesis")
	}
	if specTemplateOnly && specInteractive {
		return fmt.Errorf("--template cannot be used with --interactive")
	}

	projectRoot, cfg, setupStatus, err := resolveSpecProjectContext(promptOnly)
	if err != nil {
		return err
	}

	if promptOnly {
		return runSpecPromptOnly(args, projectRoot, cfg, outputOnly)
	}

	stop, err := runSpecSetupGate(projectRoot, cfg, setupStatus, outputOnly)
	if err != nil || stop {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)
	if err := ensureDir(specsDir); err != nil {
		return err
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

		feat, created, err = feature.EnsureExists(cfg, projectRoot, specsDir, featureRef)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	}

	if !outputOnly {
		style := styleForStdout()
		if created {
			fmt.Printf("%s %s\n", style.title("📁", "Created feature directory:"), feat.DirName)
		} else {
			fmt.Printf("%s %s\n", style.title("📁", "Using existing feature:"), feat.DirName)
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
	if changed, err := ensureSpecV2Adoption(specPath, projectRoot, feat.DirName, cfg.GoalPercentage); err != nil {
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

	if err := rollup.Update(projectRoot, cfg); err != nil {
		if !outputOnly {
			fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		}
	} else if !outputOnly {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	if !outputOnly {
		style := styleForStdout()
		fmt.Printf("\n%s %s\n", style.title("✅", "Feature ready:"), feat.Slug)
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
	}

	if needsThesisInput {
		return outputCompiledPrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg, compiledAnswers, outputOnly)
	}

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

package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

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

	if brainstormPrepare {
		if promptOnly {
			return fmt.Errorf("--prepare cannot be used with --prompt-only")
		}
		if brainstormUseVim || brainstormEditor != "" || brainstormInline || brainstormBacklog {
			return fmt.Errorf("--prepare cannot be used with --vim, --editor, --inline, or --backlog")
		}
		if brainstormOutput != "" || outputOnly || brainstormCopy {
			return fmt.Errorf("--prepare does not output a prompt; remove --output, --output-only, and --copy")
		}
		if len(args) != 1 {
			return fmt.Errorf("--prepare requires a feature name")
		}
		result, err := scaffoldBrainstormWorkflow(args[0])
		if err != nil {
			return err
		}
		return printScaffoldWorkflowResult(cmd.OutOrStdout(), "brainstorm", result)
	}

	if promptOnly {
		if brainstormUseVim || brainstormEditor != "" || brainstormInline || brainstormBacklog {
			return fmt.Errorf("--prompt-only cannot be used with --vim, --editor, --inline, or --backlog")
		}
		return outputExistingBrainstormPrompt(args, projectRoot, cfg, outputOnly)
	}

	if brainstormInline && (brainstormUseVim || brainstormEditor != "") {
		return fmt.Errorf("--inline cannot be used with --vim or --editor")
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

	feat, created, err := feature.EnsureExists(cfg, projectRoot, specsDir, featureRef)
	if err != nil {
		return err
	}
	feature.ApplyLifecycleState(feat, cfg)

	_, notesRelPath, err := ensureFeatureNotesDir(projectRoot, feat.DirName)
	if err != nil {
		return err
	}
	frontendProfileActive := effectivePromptProfile(feat.Path) == promptProfileFrontend
	if frontendProfileActive {
		if _, _, err := ensureFeatureDesignMaterialsDirs(projectRoot, feat.DirName); err != nil {
			return err
		}
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if !document.Exists(brainstormPath) {
		references := []document.MetadataReference{featureNotesReference(notesRelPath)}
		content := templates.BuildBrainstormArtifactForFeature(
			thesis,
			document.FeatureMetadataFromDir(feat.DirName),
			references,
		)
		if frontendProfileActive {
			content = seedFrontendProfileDependencyRows(content, document.TypeBrainstorm, feat.DirName)
		}
		if err := document.Write(brainstormPath, content); err != nil {
			return fmt.Errorf("failed to create BRAINSTORM.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created BRAINSTORM.md")
		}
	} else {
		if _, err := ensureBrainstormNotesDependency(brainstormPath, notesRelPath); err != nil {
			return err
		}
		if frontendProfileActive {
			if _, err := ensureFrontendProfileDependencyRows(brainstormPath, document.TypeBrainstorm, feat.DirName); err != nil {
				return err
			}
		}
		if !outputOnly {
			fmt.Println("  ✓ BRAINSTORM.md already exists")
		}
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
	preparedPrompt := prepareAgentPromptForFeature(prompt, feat.Path)

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

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/promptdoc"
	"github.com/jamesonstone/kit/internal/templates"
)

var initCopy bool
var initOutputOnly bool
var initRefresh bool
var initForce bool
var initDryRun bool
var initDiff bool
var initRefreshFiles []string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Kit project",
	Long: `Initialize a new Kit project in the current directory.

Creates:
  - .kit.yaml configuration file
  - .gitignore entries for Kit-local environment and generated artifacts
  - .env and .envrc local environment files
  - .coderabbit.yaml review configuration file
  - .github/pull_request_template.md pull request template
  - .github/workflows/auto-assign.yml issue and pull request assignment workflow
  - ~/.config/kit/.kit.yaml global configuration file
  - docs/CONSTITUTION.md
  - Repository instruction files (AGENTS.md, CLAUDE.md, .github/copilot-instructions.md)
  - Registry-managed rulesets from the Kit GitHub registry

If files already exist, Kit preserves them. Kit-managed markdown documents may
be merged by adding missing required sections.

Modes:
  Default:        Copy the prepared CONSTITUTION.md prompt to the clipboard and show next steps
  --refresh:      Refresh Kit-managed project files for an existing Kit project

Flags:
  --output-only:  Output the raw prompt to stdout instead of copying it to the clipboard
  --copy:         Copy prompt to clipboard even with --output-only
  --dry-run:      Preview --refresh without writing files
  --diff:         Print planned --refresh changes as a unified diff with --dry-run
  --force:        Overwrite refreshable generated docs during --refresh and copy a documentation review prompt
  --file:         Limit --refresh to one Kit-managed file; repeat for multiple files`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	initCmd.Flags().BoolVar(&initOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	initCmd.Flags().BoolVar(&initRefresh, "refresh", false, "refresh Kit-managed project files instead of creating a new-project prompt")
	initCmd.Flags().BoolVar(&initDryRun, "dry-run", false, "preview --refresh without writing files")
	initCmd.Flags().BoolVar(&initDiff, "diff", false, "print planned --refresh changes as a unified diff with --dry-run")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "overwrite refreshable generated docs during --refresh")
	initCmd.Flags().StringArrayVar(&initRefreshFiles, "file", nil, "limit --refresh to a Kit-managed file; repeat for multiple files")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	if initForce && !initRefresh {
		return fmt.Errorf("--force requires --refresh")
	}
	if initDryRun && !initRefresh {
		return fmt.Errorf("--dry-run requires --refresh")
	}
	if initDiff && !initRefresh {
		return fmt.Errorf("--diff requires --refresh")
	}
	if initDiff && !initDryRun {
		return fmt.Errorf("--diff requires --dry-run")
	}
	if len(initRefreshFiles) > 0 && !initRefresh {
		return fmt.Errorf("--file requires --refresh")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if initRefresh {
		return runInitRefresh(cwd, initRefreshOptions{
			force:      initForce,
			dryRun:     initDryRun,
			diff:       initDiff,
			files:      initRefreshFiles,
			outputOnly: initOutputOnly,
		})
	}

	if !initOutputOnly {
		fmt.Println("🎒 Initializing Kit project...")
	}

	// create or merge .kit.yaml
	cfg := defaultInitConfig()
	if config.Exists(cwd) {
		if !initOutputOnly {
			fmt.Println("  ✓ .kit.yaml exists, merging...")
		}
		existing, err := config.Load(cwd)
		if err == nil {
			cfg = existing
		}
		if !config.IsInstructionScaffoldVersionSupported(cfg.InstructionScaffoldVersion) {
			cfg.InstructionScaffoldVersion = detectInstructionScaffoldVersion(cwd, cfg)
			if cfg.InstructionScaffoldVersion == instructionScaffoldVersionUnknown {
				cfg.InstructionScaffoldVersion = config.DefaultInstructionScaffoldVersion
			}
			if err := config.Save(cwd, cfg); err != nil {
				return fmt.Errorf("failed to update %s: %w", config.ConfigFileName, err)
			}
		}
	} else {
		if err := config.Save(cwd, cfg); err != nil {
			return fmt.Errorf("failed to create .kit.yaml: %w", err)
		}
		if !initOutputOnly {
			fmt.Println("  ✓ Created .kit.yaml")
		}
	}

	if err := populateGlobalConfig(initOutputOnly); err != nil {
		return err
	}
	if err := scaffoldGitignore(cwd, initOutputOnly); err != nil {
		return err
	}
	if err := scaffoldEnvFiles(cwd, initOutputOnly); err != nil {
		return err
	}
	if err := scaffoldCodeRabbitConfig(cwd, initOutputOnly); err != nil {
		return err
	}
	if err := scaffoldPullRequestTemplate(cwd, initOutputOnly); err != nil {
		return err
	}
	if err := scaffoldAutoAssignWorkflow(cwd, cfg, initOutputOnly); err != nil {
		return err
	}

	// ensure docs directory exists
	docsDir := filepath.Join(cwd, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// ensure specs directory exists
	specsDir := cfg.SpecsPath(cwd)
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		return fmt.Errorf("failed to create specs directory: %w", err)
	}
	if !initOutputOnly {
		fmt.Println("  ✓ Created docs/specs/")
	}

	// create or merge CONSTITUTION.md
	constitutionPath := cfg.ConstitutionAbsPath(cwd)
	if document.Exists(constitutionPath) {
		if !initOutputOnly {
			fmt.Println("  ✓ docs/CONSTITUTION.md exists, merging...")
		}
		if err := document.MergeDocument(constitutionPath, templates.Constitution, document.TypeConstitution); err != nil {
			return fmt.Errorf("failed to merge CONSTITUTION.md: %w", err)
		}
	} else {
		if err := document.Write(constitutionPath, templates.Constitution); err != nil {
			return fmt.Errorf("failed to create CONSTITUTION.md: %w", err)
		}
		if !initOutputOnly {
			fmt.Println("  ✓ Created docs/CONSTITUTION.md")
		}
	}

	// scaffold repository instruction files
	for _, relativePath := range instructionArtifactPaths(
		cfg,
		instructionFileSelection{},
		cfg.InstructionScaffoldVersion,
		true,
	) {
		result, err := writeInstructionFileWithMode(
			cwd,
			relativePath,
			instructionFileWriteModeSkipExisting,
			cfg.InstructionScaffoldVersion,
		)
		if err != nil {
			return err
		}

		switch result {
		case instructionFileCreated:
			if !initOutputOnly {
				fmt.Printf("  ✓ Created %s\n", relativePath)
			}
		case instructionFileSkipped:
			if !initOutputOnly {
				fmt.Printf("  ✓ %s exists, skipping\n", relativePath)
			}
		}
	}

	if err := runInitRefresh(cwd, initRefreshOptions{outputOnly: true}); err != nil {
		return err
	}

	if !initOutputOnly {
		fmt.Println("\n✅ Kit project initialized!")
	}

	// output easy-to-copy instruction for coding agents
	constitutionRelPath := cfg.ConstitutionPath
	constitutionFullPath := filepath.Join(cwd, constitutionRelPath)
	prompt := buildProjectInitPrompt(cwd, constitutionFullPath)

	if err := outputPromptWithClipboardDefault(prompt, initOutputOnly, initCopy); err != nil {
		return err
	}

	if !initOutputOnly {
		printNumberedNextSteps([]string{
			"Paste the copied prompt into your agent to draft docs/CONSTITUTION.md",
			"Review and refine docs/CONSTITUTION.md to define project constraints",
			"Run `kit spec <feature-name>` to create your first feature",
		})
	}

	return nil
}

func defaultInitConfig() *config.Config {
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.DefaultInstructionScaffoldVersion
	ensureInitLoopReviewConfig(cfg)
	return cfg
}

func populateGlobalConfig(outputOnly bool) error {
	configPath, changed, err := config.PopulateGlobalConfig(defaultInitConfig())
	if err != nil {
		return fmt.Errorf("failed to populate global config: %w", err)
	}

	if outputOnly {
		return nil
	}
	if changed {
		fmt.Printf("  ✓ Populated %s\n", configPath)
		return nil
	}
	fmt.Printf("  ✓ %s exists\n", configPath)
	return nil
}

func buildProjectInitPrompt(projectRoot, constitutionFullPath string) string {
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf(
			"Please update %s with all patterns, strategy, implementation details, process, and long-term vision for this project.\nThis document will drive the \"rules for development\" going forward.",
			constitutionFullPath,
		))
		doc.Paragraph(fmt.Sprintf("Analyze the codebase at %s to extract:", projectRoot))
		doc.BulletList(
			"Architectural patterns and conventions",
			"Code style and naming conventions",
			"Dependencies and their purposes",
			"Non-negotiable constraints",
			"Project goals and non-goals",
		)
		doc.Paragraph("Rules:")
		doc.BulletList("PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times")
	})
}

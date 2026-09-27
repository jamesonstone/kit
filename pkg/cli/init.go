package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

var initCopy bool
var initOutputOnly bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Kit project",
	Long: `Initialize a new Kit project in the current directory.

Creates:
  - .kit.yaml configuration file
  - .gitignore entries for Kit-local environment and generated artifacts
  - .env and .envrc local environment files
  - Makefile starter for canonical project commands
  - .coderabbit.yaml review configuration file
  - .github/pull_request_template.md pull request template
  - .github/workflows/auto-assign.yml issue and pull request assignment workflow
  - README.md Kit-managed status badges and final Maintainers section
  - ~/.config/kit/.kit.yaml global configuration file
  - docs/CONSTITUTION.md
  - Repository instruction files (AGENTS.md, CLAUDE.md, .github/copilot-instructions.md)
  - The core rules shipped with this Kit binary

Init is for projects that are not yet Kit projects. Existing files are
preserved. In a project that already has .kit.yaml, run ` + "`kit reconcile`" + `
instead: it brings any existing Kit project to the current structure.

The generated Constitution starter is a valid bootstrap state. The prepared
prompt promotes only durable project-wide truth supported by repository evidence
and leaves empty-project Constitution sections unchanged.

Init copies the prepared project initialization prompt to the clipboard and
shows next steps; --output-only prints it instead.`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	initCmd.Flags().BoolVar(&initOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if config.Exists(cwd) {
		return fmt.Errorf("%s already exists, so this is a Kit project; run `kit reconcile` to bring it to the current structure", config.ConfigFileName)
	}

	deliveryCfg := defaultInitConfig()
	deliveryBaseline, err := captureManagedFileDeliveryBaseline(
		cwd,
		projectInitDeliveryPaths(deliveryCfg),
	)
	if err != nil {
		return err
	}

	if !initOutputOnly {
		fmt.Println("🎒 Initializing Kit project...")
	}

	cfg := defaultInitConfig()
	if err := config.Save(cwd, cfg); err != nil {
		return fmt.Errorf("failed to create .kit.yaml: %w", err)
	}
	if !initOutputOnly {
		fmt.Println("  ✓ Created .kit.yaml")
	}
	if !initOutputOnly && streamsHaveInteractiveTerminal(os.Stdin, os.Stdout) {
		inspectionCfg, inspection, err := config.LoadWithInspection(cwd)
		if err != nil {
			return err
		}
		changed, err := remediateProjectConfig(cwd, inspectionCfg, inspection, configRemediationOptions{
			Interactive: true,
			Input:       os.Stdin,
			Output:      os.Stdout,
		})
		if err != nil {
			return err
		}
		if changed {
			cfg, err = config.Load(cwd)
			if err != nil {
				return err
			}
		}
	}

	if err := populateGlobalConfig(initOutputOnly); err != nil {
		return err
	}
	if err := scaffoldGitignore(cwd, initOutputOnly); err != nil {
		return err
	}
	if err := scaffoldMakefile(cwd, initOutputOnly); err != nil {
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
	for _, relativePath := range instructionArtifactPaths(cfg) {
		result, err := writeInstructionFileWithMode(cwd, relativePath, instructionFileWriteModeSkipExisting)
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

	refreshDeliverySnapshot, err := runInitRefreshWithSnapshot(cwd, initRefreshOptions{outputOnly: true})
	if err != nil {
		return err
	}
	initialDeliverySnapshot, err := managedFileDeliverySnapshotFromBaseline(cwd, deliveryBaseline)
	if err != nil {
		return err
	}
	deliverySnapshot := mergeManagedFileDeliverySnapshots(
		initialDeliverySnapshot,
		refreshDeliverySnapshot,
	)

	if !initOutputOnly {
		fmt.Println("\n✅ Kit project initialized!")
	}

	// output easy-to-copy instruction for coding agents
	constitutionRelPath := cfg.ConstitutionPath
	constitutionFullPath := filepath.Join(cwd, constitutionRelPath)
	prompt := buildProjectInitPrompt(cwd, constitutionFullPath, deliverySnapshot)

	if err := outputPromptWithClipboardDefault(prompt, initOutputOnly, initCopy); err != nil {
		return err
	}

	if !initOutputOnly {
		printNumberedNextSteps([]string{
			"Paste the copied prompt into your agent to review repository evidence and populate only verified Makefile targets",
			"Keep the starter Constitution unchanged until implemented evidence supports durable project-wide rules",
			"Run `kit spec <feature-name>` to create your first feature",
		})
	}

	return nil
}

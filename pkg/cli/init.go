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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Kit project",
	Long: `Initialize a new Kit project in the current directory.

Creates:
  - .kit.yaml configuration file
  - docs/CONSTITUTION.md
	- Repository instruction files (AGENTS.md, CLAUDE.md, .github/copilot-instructions.md)

If files already exist, Kit attempts to merge by preserving existing
content and adding any missing required sections.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	fmt.Println("🎒 Initializing Kit project...")

	// create or merge .kit.yaml
	cfg := config.Default()
	if config.Exists(cwd) {
		fmt.Println("  ✓ .kit.yaml exists, merging...")
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
		cfg.InstructionScaffoldVersion = config.DefaultInstructionScaffoldVersion
		if err := config.Save(cwd, cfg); err != nil {
			return fmt.Errorf("failed to create .kit.yaml: %w", err)
		}
		fmt.Println("  ✓ Created .kit.yaml")
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
	fmt.Println("  ✓ Created docs/specs/")

	// create or merge CONSTITUTION.md
	constitutionPath := cfg.ConstitutionAbsPath(cwd)
	if document.Exists(constitutionPath) {
		fmt.Println("  ✓ docs/CONSTITUTION.md exists, merging...")
		if err := document.MergeDocument(constitutionPath, templates.Constitution, document.TypeConstitution); err != nil {
			return fmt.Errorf("failed to merge CONSTITUTION.md: %w", err)
		}
	} else {
		if err := document.Write(constitutionPath, templates.Constitution); err != nil {
			return fmt.Errorf("failed to create CONSTITUTION.md: %w", err)
		}
		fmt.Println("  ✓ Created docs/CONSTITUTION.md")
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
			fmt.Printf("  ✓ Created %s\n", relativePath)
		case instructionFileSkipped:
			fmt.Printf("  ✓ %s exists, skipping\n", relativePath)
		}
	}

	fmt.Println("\n✅ Kit project initialized!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit docs/CONSTITUTION.md to define project constraints")
	fmt.Println("  2. Run 'kit spec <feature-name>' to create your first feature")

	// output easy-to-copy instruction for coding agents
	constitutionRelPath := cfg.ConstitutionPath
	constitutionFullPath := filepath.Join(cwd, constitutionRelPath)

	prompt := renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf(
			"Please update %s with all patterns, strategy, implementation details, process, and long-term vision for this project.\nThis document will drive the \"rules for development\" going forward.",
			constitutionFullPath,
		))
		doc.Paragraph(fmt.Sprintf("Analyze the codebase at %s to extract:", cwd))
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

	if err := outputPrompt(prompt, false, false); err != nil {
		return err
	}

	return nil
}

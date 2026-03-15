package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

var scaffoldAgentsForce bool
var scaffoldAgentsCopilot bool
var scaffoldAgentsClaude bool
var scaffoldAgentsAgentsMD bool

var scaffoldAgentsCmd = &cobra.Command{
	Use:     "scaffold-agents",
	Aliases: []string{"scaffold-agent"},
	Short:   "Create or refresh repository instruction files",
	Long: `Create missing repository instruction files and optionally overwrite existing ones.

Repository instruction files include:
	- AGENTS.md
	- CLAUDE.md
	- .github/copilot-instructions.md

These files contain:
  - Links to canonical documents
  - Workflow contracts for each agent
	- Change classification and execution rules
	- Shared quality gates and coding standards

These files stay aligned with canonical project documents.

Use --agentsmd, --claude, and --copilot to update only specific built-in files.

Use --force to overwrite existing files.`,
	RunE: runScaffoldAgents,
}

func init() {
	scaffoldAgentsCmd.Flags().BoolVarP(&scaffoldAgentsForce, "force", "f", false, "overwrite existing repository instruction files")
	scaffoldAgentsCmd.Flags().BoolVar(&scaffoldAgentsAgentsMD, "agentsmd", false, "scaffold only AGENTS.md")
	scaffoldAgentsCmd.Flags().BoolVar(&scaffoldAgentsClaude, "claude", false, "scaffold only CLAUDE.md")
	scaffoldAgentsCmd.Flags().BoolVar(&scaffoldAgentsCopilot, "copilot", false, "scaffold only .github/copilot-instructions.md")
	rootCmd.AddCommand(scaffoldAgentsCmd)
}

func runScaffoldAgents(cmd *cobra.Command, args []string) error {
	selection := instructionFileSelection{
		agentsMD: scaffoldAgentsAgentsMD,
		claude:   scaffoldAgentsClaude,
		copilot:  scaffoldAgentsCopilot,
	}

	projectRoot, cfg, err := scaffoldInstructionContext(selection)
	if err != nil {
		return err
	}

	fmt.Println("🤖 Scaffolding repository instruction files...")

	var created, updated, skipped int

	for _, instructionFile := range selectedInstructionFiles(cfg, selection) {
		result, err := writeInstructionFile(projectRoot, instructionFile, scaffoldAgentsForce)
		if err != nil {
			return err
		}

		switch result {
		case instructionFileCreated:
			fmt.Printf("  ✓ Created %s\n", instructionFile)
			created++
		case instructionFileUpdated:
			fmt.Printf("  ✓ Overwrote %s\n", instructionFile)
			updated++
		case instructionFileSkipped:
			fmt.Printf("  ✓ %s exists (skipped)\n", instructionFile)
			skipped++
		}
	}

	fmt.Printf("\n✅ Instruction scaffolding complete!\n")
	fmt.Printf("   Created: %d, Updated: %d, Skipped: %d\n", created, updated, skipped)

	return nil
}

func scaffoldInstructionContext(selection instructionFileSelection) (string, *config.Config, error) {
	projectRoot, err := config.FindProjectRoot()
	if err == nil {
		cfg, err := config.Load(projectRoot)
		if err != nil {
			return "", nil, err
		}

		return projectRoot, cfg, nil
	}

	if !selection.any() {
		return "", nil, err
	}

	projectRoot, cwdErr := os.Getwd()
	if cwdErr != nil {
		return "", nil, fmt.Errorf("failed to get working directory: %w", cwdErr)
	}

	return projectRoot, config.Default(), nil
}

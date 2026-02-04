package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

var agentsMDCmd = &cobra.Command{
	Use:   "agentsmd",
	Short: "Create or overwrite AGENTS.md with the comprehensive template",
	Long: `Create or overwrite AGENTS.md in the project root.

This command writes the full AGENTS.md template containing:
  - Kit source of truth references
  - Communication style guidelines
  - Plan → Act → Reflect workflow
  - Definition of Done
  - Code style and architecture standards
  - Testing, logging, and security guidelines
  - Git rules and document management

If AGENTS.md already exists, it will be overwritten.`,
	RunE: runAgentsMD,
}

func init() {
	rootCmd.AddCommand(agentsMDCmd)
}

func runAgentsMD(cmd *cobra.Command, args []string) error {
	// find project root
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		// fallback to cwd if not in a kit project
		projectRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	agentsPath := filepath.Join(projectRoot, "AGENTS.md")

	// check if file exists for messaging
	exists := false
	if _, err := os.Stat(agentsPath); err == nil {
		exists = true
	}

	// write the file (create or overwrite)
	if err := os.WriteFile(agentsPath, []byte(templates.AgentsMD), 0644); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	if exists {
		fmt.Println("✅ Overwrote AGENTS.md")
	} else {
		fmt.Println("✅ Created AGENTS.md")
	}

	return nil
}

package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

var scaffoldAgentsForce bool

var scaffoldAgentsCmd = &cobra.Command{
	Use:   "scaffold-agents",
	Short: "Create or update agent pointer files",
	Long: `Create missing agent pointer files and update document links.

Agent pointer files (e.g., AGENTS.md, CLAUDE.md) contain:
  - Links to canonical documents
  - Workflow contracts for each agent
  - Multi-feature rules

These files never duplicate specifications; they only point to them.

Use --force to overwrite existing files.`,
	RunE: runScaffoldAgents,
}

func init() {
	scaffoldAgentsCmd.Flags().BoolVarP(&scaffoldAgentsForce, "force", "f", false, "overwrite existing agent pointer files")
	rootCmd.AddCommand(scaffoldAgentsCmd)
}

func runScaffoldAgents(cmd *cobra.Command, args []string) error {
	// find project root
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	fmt.Println("🤖 Scaffolding agent pointer files...")

	var created, updated, skipped int

	for _, agentFile := range cfg.Agents {
		agentPath := filepath.Join(projectRoot, agentFile)
		agentName := agentFile[:len(agentFile)-3] // remove .md extension

		if document.Exists(agentPath) && !scaffoldAgentsForce {
			fmt.Printf("  ✓ %s exists (skipped)\n", agentFile)
			skipped++
			continue
		}

		existed := document.Exists(agentPath)
		content := templates.AgentPointer(agentName)
		if err := document.Write(agentPath, content); err != nil {
			return fmt.Errorf("failed to write %s: %w", agentFile, err)
		}

		if existed {
			fmt.Printf("  ✓ Overwrote %s\n", agentFile)
			updated++
		} else {
			fmt.Printf("  ✓ Created %s\n", agentFile)
			created++
		}
	}

	fmt.Printf("\n✅ Agent scaffolding complete!\n")
	fmt.Printf("   Created: %d, Updated: %d, Skipped: %d\n", created, updated, skipped)

	return nil
}

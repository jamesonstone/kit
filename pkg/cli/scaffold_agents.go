package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

var scaffoldAgentsForce bool
var scaffoldAgentsCopilot bool
var scaffoldAgentsClaude bool
var scaffoldAgentsAgentsMD bool
var scaffoldAgentsYes bool
var scaffoldAgentsAppendOnly bool
var scaffoldAgentsVersion int

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

Use --force to overwrite existing files.
Use --append-only to add missing Kit-managed sections without overwriting
existing matched content.`,
	RunE: runScaffoldAgents,
}

func init() {
	scaffoldAgentsCmd.Flags().BoolVarP(&scaffoldAgentsForce, "force", "f", false, "overwrite existing repository instruction files")
	scaffoldAgentsCmd.Flags().BoolVarP(&scaffoldAgentsYes, "yes", "y", false, "skip the overwrite confirmation prompt when used with --force")
	scaffoldAgentsCmd.Flags().BoolVar(&scaffoldAgentsAppendOnly, "append-only", false, "append missing Kit-managed sections without overwriting matched existing content")
	scaffoldAgentsCmd.Flags().BoolVar(&scaffoldAgentsAgentsMD, "agentsmd", false, "scaffold only AGENTS.md")
	scaffoldAgentsCmd.Flags().BoolVar(&scaffoldAgentsClaude, "claude", false, "scaffold only CLAUDE.md")
	scaffoldAgentsCmd.Flags().BoolVar(&scaffoldAgentsCopilot, "copilot", false, "scaffold only .github/copilot-instructions.md")
	scaffoldAgentsCmd.Flags().IntVar(&scaffoldAgentsVersion, "version", 0, "instruction scaffold version: 1 = verbose legacy, 2 = thin ToC/RLM model (default)")
	rootCmd.AddCommand(scaffoldAgentsCmd)
}

func runScaffoldAgents(cmd *cobra.Command, args []string) error {
	if scaffoldAgentsYes && !scaffoldAgentsForce {
		return fmt.Errorf("--yes requires --force")
	}

	targetVersion, versionExplicit, err := resolveInstructionScaffoldVersionFlag(scaffoldAgentsVersion)
	if err != nil {
		return err
	}

	writeMode, err := determineInstructionFileWriteMode(scaffoldAgentsForce, scaffoldAgentsAppendOnly)
	if err != nil {
		return err
	}

	selection := instructionFileSelection{
		agentsMD: scaffoldAgentsAgentsMD,
		claude:   scaffoldAgentsClaude,
		copilot:  scaffoldAgentsCopilot,
	}

	projectRoot, cfg, err := scaffoldInstructionContext(selection)
	if err != nil {
		return err
	}
	currentVersion := detectInstructionScaffoldVersion(projectRoot, cfg)
	if !versionExplicit {
		if currentVersion != instructionScaffoldVersionUnknown {
			targetVersion = currentVersion
		} else {
			targetVersion = config.DefaultInstructionScaffoldVersion
		}
	}
	forceFullModel := instructionVersionChangeRequiresForce(currentVersion, targetVersion)
	if forceFullModel && writeMode != instructionFileWriteModeOverwrite {
		return fmt.Errorf(
			"switching the instruction scaffold from version %d to version %d requires --force. Re-run `kit scaffold-agents --version %d --force` to confirm the repo-wide change",
			currentVersion,
			targetVersion,
			targetVersion,
		)
	}

	fmt.Printf("🤖 Scaffolding repository instruction files (version %d)...\n", targetVersion)
	targets := instructionArtifactPaths(cfg, selection, targetVersion, forceFullModel)

	if writeMode == instructionFileWriteModeOverwrite && !scaffoldAgentsYes {
		existing := existingInstructionFiles(projectRoot, targets)
		if len(existing) > 0 {
			ok, err := confirmInstructionOverwrite(cmd, existing)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("\nCancelled. No instruction files were changed.")
				return nil
			}
		}
	}

	cleanupPlans, err := planInstructionVersionCleanup(projectRoot, currentVersion, targetVersion)
	if err != nil {
		return err
	}

	plans, err := planInstructionArtifactWrites(projectRoot, targets, writeMode, targetVersion)
	if err != nil {
		return err
	}

	var created, updated, merged, skipped, removed int

	for _, plan := range plans {
		result, err := applyInstructionFileWritePlan(plan)
		if err != nil {
			return err
		}

		switch result {
		case instructionFileCreated:
			fmt.Printf("  ✓ Created %s\n", plan.relativePath)
			created++
		case instructionFileUpdated:
			fmt.Printf("  ✓ Overwrote %s\n", plan.relativePath)
			updated++
		case instructionFileMerged:
			fmt.Printf("  ✓ Appended missing Kit sections to %s\n", plan.relativePath)
			merged++
		case instructionFileSkipped:
			if writeMode == instructionFileWriteModeAppendOnly {
				fmt.Printf("  ✓ %s already contains all detectable Kit sections\n", plan.relativePath)
			} else {
				fmt.Printf("  ✓ %s exists (skipped)\n", plan.relativePath)
			}
			skipped++
		}
	}

	if len(cleanupPlans) > 0 {
		cleanupRemoved, err := applyInstructionVersionCleanup(projectRoot, cleanupPlans)
		if err != nil {
			return err
		}
		removed += cleanupRemoved
	}

	cfg.InstructionScaffoldVersion = targetVersion
	if err := config.Save(projectRoot, cfg); err != nil {
		return err
	}

	fmt.Printf("\n✅ Instruction scaffolding complete!\n")
	fmt.Printf(
		"   Created: %d, Updated: %d, Merged: %d, Removed: %d, Skipped: %d\n",
		created,
		updated,
		merged,
		removed,
		skipped,
	)

	if writeMode == instructionFileWriteModeSkipExisting && skipped > 0 {
		fmt.Println("   Hint: use --append-only to merge missing Kit-managed sections without overwriting custom content, or --force to replace existing files.")
	}

	return nil
}

func confirmInstructionOverwrite(cmd *cobra.Command, existingFiles []string) (bool, error) {
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintln(out, "The following repository instruction files will be overwritten:"); err != nil {
		return false, err
	}
	for _, file := range existingFiles {
		if _, err := fmt.Fprintf(out, "- %s\n", file); err != nil {
			return false, err
		}
	}
	if _, err := fmt.Fprintln(out, "Proceed? [y/N]"); err != nil {
		return false, err
	}

	input, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "y" || answer == "yes", nil
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

package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var (
	ciPRRef       string
	ciRunID       string
	ciJobRef      string
	ciWorkflowRef string
	ciRepoPath    string
	ciJSON        bool
	ciDispatch    bool
	ciUseCopilot  bool
	ciNoCopilot   bool
	ciLogLines    int
	ciUseVim      bool
	ciEditor      string
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Diagnose GitHub Actions failures",
	Long: `Diagnose GitHub Actions failures for the current repository.

By default, kit ci inspects the latest failed run on the discovered default
branch. It can also target a pull request, workflow run, workflow file/name, or
one job within a workflow run. The command is diagnostic: it does not edit
source files, rerun CI, commit, push, or mutate GitHub state.`,
	Args: cobra.NoArgs,
	RunE: runCI,
}

func init() {
	addFreeTextInputFlags(ciCmd, &ciUseVim, &ciEditor)
	ciUseCopilot = true
	ciCmd.Flags().StringVar(&ciPRRef, "pr", "", "diagnose checks for a PR number, URL, Markdown link, or owner/repo#number")
	ciCmd.Flags().StringVar(&ciRunID, "run", "", "diagnose a specific GitHub Actions run ID")
	ciCmd.Flags().StringVar(&ciJobRef, "job", "", "with --run, diagnose a specific job name or job database ID")
	ciCmd.Flags().StringVar(&ciWorkflowRef, "workflow", "", "diagnose the latest failure for a workflow name or path")
	ciCmd.Flags().StringVar(&ciRepoPath, "repo-path", "", "run the diagnosis against a specific local repository path")
	ciCmd.Flags().BoolVar(&ciJSON, "json", false, "print structured JSON")
	ciCmd.Flags().BoolVar(&ciDispatch, "dispatch", false, "open the dispatch editor prefilled with CI diagnosis context")
	ciCmd.Flags().BoolVar(&ciUseCopilot, "copilot", true, "attempt callable GitHub Copilot diagnosis when available")
	ciCmd.Flags().BoolVar(&ciNoCopilot, "no-copilot", false, "disable GitHub Copilot diagnosis attempts")
	ciCmd.Flags().IntVar(&ciLogLines, "log-lines", 200, "maximum relevant log lines to include per failed job")
	rootCmd.AddCommand(ciCmd)
}

func runCI(cmd *cobra.Command, args []string) error {
	opts := ciOptions{
		PRRef:       ciPRRef,
		RunID:       ciRunID,
		JobRef:      ciJobRef,
		WorkflowRef: ciWorkflowRef,
		RepoPath:    ciRepoPath,
		JSON:        ciJSON,
		Dispatch:    ciDispatch,
		UseCopilot:  ciUseCopilot && !ciNoCopilot,
		NoCopilot:   ciNoCopilot,
		LogLines:    ciLogLines,
		InputConfig: newFreeTextInputConfig(ciUseVim, ciEditor, false, true),
	}
	if ciUseCopilot && ciNoCopilot &&
		cmd.Flags().Changed("copilot") &&
		cmd.Flags().Changed("no-copilot") {
		return newCLIExitError(fmt.Errorf("--copilot and --no-copilot cannot both be set"), 2, false)
	}

	exitCode, err := runCIWithOptions(opts, cmd.OutOrStdout())
	if err != nil {
		return newCLIExitError(err, 2, false)
	}
	if exitCode != 0 {
		return newCLIExitError(fmt.Errorf("CI failure diagnosed"), exitCode, true)
	}
	return nil
}

func runCIWithOptions(opts ciOptions, out io.Writer) (int, error) {
	if opts.LogLines <= 0 {
		return 0, fmt.Errorf("--log-lines must be greater than 0")
	}

	diagnosis, err := buildCIDiagnosis(opts)
	if err != nil {
		return 0, err
	}

	if opts.JSON {
		if err := renderCIDiagnosisJSON(out, diagnosis); err != nil {
			return 0, err
		}
	} else {
		if err := renderCIDiagnosisHuman(out, diagnosis); err != nil {
			return 0, err
		}
	}

	if opts.Dispatch && diagnosis.FailureFound {
		if err := openCIDispatchPrompt(opts, diagnosis); err != nil {
			return 0, err
		}
	}

	if diagnosis.FailureFound {
		return 1, nil
	}
	return 0, nil
}

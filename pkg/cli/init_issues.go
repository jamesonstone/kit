package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

func bindIssueProviderFlags(cmd *cobra.Command) {
	cmd.Flags().String("issue-provider", "github", "development issue provider: github or linear (repository-local)")
	cmd.Flags().String("linear-workspace", "", "verified Linear workspace URL slug (not a credential)")
	cmd.Flags().String("linear-team-id", "", "verified Linear team UUID")
	cmd.Flags().String("linear-project-id", "", "verified Linear project UUID; omit for team-only work")
}

func selectInitIssueProvider(cmd *cobra.Command, cfg *config.Config) error {
	provider, _ := cmd.Flags().GetString("issue-provider")
	workspace, _ := cmd.Flags().GetString("linear-workspace")
	team, _ := cmd.Flags().GetString("linear-team-id")
	project, _ := cmd.Flags().GetString("linear-project-id")
	selected := config.IssueConfig{Provider: provider, Linear: config.LinearConfig{Workspace: workspace, TeamID: team, ProjectID: project}}
	if !cmd.Flags().Changed("issue-provider") && !initOutputOnly && streamsHaveInteractiveTerminal(cmd.InOrStdin(), cmd.OutOrStdout()) {
		var err error
		selected, err = promptIssueProvider(cmd.InOrStdin(), cmd.ErrOrStderr(), selected)
		if err != nil {
			return err
		}
	}
	if err := selected.Validate(); err != nil {
		return err
	}
	if selected.EffectiveProvider() == "github" && (selected.Linear != (config.LinearConfig{})) {
		return fmt.Errorf("linear target flags require --issue-provider linear")
	}
	if selected.Provider == "github" {
		selected.Provider = ""
	} // preserve legacy omitted-default YAML
	cfg.Issues = selected
	return nil
}

func promptIssueProvider(input io.Reader, out io.Writer, selected config.IssueConfig) (config.IssueConfig, error) {
	reader := bufio.NewReader(input)
	read := func(prompt, fallback string) (string, error) {
		if _, err := fmt.Fprint(out, prompt); err != nil {
			return "", err
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read issue provider selection: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			line = fallback
		}
		return line, nil
	}
	var err error
	selected.Provider, err = read("Issue provider: [1] GitHub Issues (default), [2] Linear: ", "github")
	if err != nil {
		return selected, err
	}
	switch selected.Provider {
	case "1":
		selected.Provider = "github"
	case "2":
		selected.Provider = "linear"
	}
	if selected.Provider != "linear" {
		return selected, selected.Validate()
	}
	selected.Linear.Workspace, err = read("Verified Linear workspace URL slug: ", selected.Linear.Workspace)
	if err != nil {
		return selected, err
	}
	selected.Linear.TeamID, err = read("Verified Linear team UUID: ", selected.Linear.TeamID)
	if err != nil {
		return selected, err
	}
	selected.Linear.ProjectID, err = read("Verified Linear project UUID (empty for team-only work): ", selected.Linear.ProjectID)
	if err != nil {
		return selected, err
	}
	return selected, selected.Validate()
}

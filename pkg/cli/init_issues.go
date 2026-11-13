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
	cmd.Flags().String("linear-team-key", "", "verified Linear team identifier for the branch prefix (for example ENG)")
	cmd.Flags().String("linear-project-id", "", "verified Linear project UUID; omit for team-only work")
}

func printIssueBranchPrefix(out io.Writer, prefix string) error {
	if prefix == "" {
		prefix = "resolve from the selected Linear team"
	}
	_, err := fmt.Fprintf(out, "Branch prefix: %s\n", prefix)
	return err
}

func selectInitIssueProvider(cmd *cobra.Command, cfg *config.Config) error {
	provider, _ := cmd.Flags().GetString("issue-provider")
	workspace, _ := cmd.Flags().GetString("linear-workspace")
	team, _ := cmd.Flags().GetString("linear-team-id")
	teamKey, _ := cmd.Flags().GetString("linear-team-key")
	project, _ := cmd.Flags().GetString("linear-project-id")
	selected := config.IssueConfig{Provider: provider, Linear: config.LinearConfig{
		Workspace: workspace, TeamID: team, TeamKey: teamKey, ProjectID: project,
	}}
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
	previous := selected.Linear
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
	keyPrompt := "Verified Linear team identifier (for example ENG, empty to leave unset): "
	if previous.SameTeam(selected.Linear) && previous.TeamKey != "" {
		keyPrompt = fmt.Sprintf("Verified Linear team identifier [%s]: ", previous.TeamKey)
	} else {
		selected.Linear.TeamKey = ""
	}
	selected.Linear.TeamKey, err = read(keyPrompt, selected.Linear.TeamKey)
	if err != nil {
		return selected, err
	}
	return selected, selected.Validate()
}

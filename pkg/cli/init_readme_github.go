package cli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

var readmeCommandRunner ciCommandRunner = execCICommandRunner{}

type readmeGitHubRepo struct {
	Repository string
	Visibility string
}

func readmeGitHubRepository(projectRoot string, cfg *config.Config) (readmeGitHubRepo, error) {
	if cfg != nil {
		if repository, ok := normalizeGitHubRepository(cfg.GitHub.Repository); ok {
			return readmeGitHubRepo{
				Repository: repository,
				Visibility: readmeGitHubVisibility(projectRoot, repository),
			}, nil
		}
	}

	output, err := readmeCommandRunner.Output(projectRoot, "git", "remote", "get-url", "origin")
	if err != nil {
		return readmeGitHubRepo{}, fmt.Errorf("failed to resolve GitHub repository for README badges: %w", err)
	}
	owner, repo, err := parseGitHubRemoteURL(strings.TrimSpace(string(output)))
	if err != nil {
		return readmeGitHubRepo{}, fmt.Errorf("failed to resolve GitHub repository for README badges: %w", err)
	}
	repository := owner + "/" + repo
	return readmeGitHubRepo{
		Repository: repository,
		Visibility: readmeGitHubVisibility(projectRoot, repository),
	}, nil
}

func readmeGitHubVisibility(projectRoot, repository string) string {
	output, err := readmeCommandRunner.Output(projectRoot, "gh", "repo", "view", repository, "--json", "visibility", "-q", ".visibility")
	if err != nil {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(string(output)))
}

func normalizeGitHubRepository(raw string) (string, bool) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(raw), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	return parts[0] + "/" + strings.TrimSuffix(parts[1], ".git"), true
}

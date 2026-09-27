package cli

import (
	"fmt"
	"regexp"
	"strings"
)

var githubRemotePattern = regexp.MustCompile(`github\.com[:/]([^/\s]+)/([^/\s]+?)(?:\.git)?$`)

func parseGitHubRemoteURL(raw string) (string, string, error) {
	if match := githubRemotePattern.FindStringSubmatch(strings.TrimSpace(raw)); match != nil {
		return match[1], strings.TrimSuffix(match[2], ".git"), nil
	}
	return "", "", fmt.Errorf("remote origin is not a GitHub repository URL: %s", raw)
}

func shellQuoteArgument(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

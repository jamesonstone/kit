package cli

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func resolveDispatchPRTarget(raw string) (dispatchPRTarget, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return dispatchPRTarget{}, fmt.Errorf("--pr cannot be empty")
	}

	if match := dispatchPRURLPattern.FindStringSubmatch(value); match != nil {
		number, err := strconv.Atoi(match[3])
		if err != nil {
			return dispatchPRTarget{}, fmt.Errorf("invalid PR number %q: %w", match[3], err)
		}
		return dispatchPRTarget{Owner: match[1], Repo: match[2], Number: number}, nil
	}

	if parsed, err := url.Parse(value); err == nil && parsed.Host == "github.com" {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) >= 4 && parts[2] == "pull" {
			number, err := strconv.Atoi(parts[3])
			if err != nil {
				return dispatchPRTarget{}, fmt.Errorf("invalid PR number %q: %w", parts[3], err)
			}
			return dispatchPRTarget{Owner: parts[0], Repo: parts[1], Number: number}, nil
		}
	}

	if match := dispatchPROwnerNumberPattern.FindStringSubmatch(value); match != nil {
		number, err := strconv.Atoi(match[3])
		if err != nil {
			return dispatchPRTarget{}, fmt.Errorf("invalid PR number %q: %w", match[3], err)
		}
		return dispatchPRTarget{Owner: match[1], Repo: match[2], Number: number}, nil
	}

	if number, err := strconv.Atoi(value); err == nil {
		owner, repo, err := dispatchCurrentRepoResolver()
		if err != nil {
			return dispatchPRTarget{}, err
		}
		return dispatchPRTarget{Owner: owner, Repo: repo, Number: number}, nil
	}

	return dispatchPRTarget{}, fmt.Errorf("could not parse PR reference %q", raw)
}

func resolveCurrentGitHubRepo() (string, string, error) {
	output, err := commandOutput("git", "remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve current repo from git remote origin: %w", err)
	}

	owner, repo, err := parseGitHubRemoteURL(strings.TrimSpace(string(output)))
	if err != nil {
		return "", "", err
	}

	return owner, repo, nil
}

func parseGitHubRemoteURL(raw string) (string, string, error) {
	if match := dispatchGitHubRemotePattern.FindStringSubmatch(strings.TrimSpace(raw)); match != nil {
		return match[1], strings.TrimSuffix(match[2], ".git"), nil
	}

	return "", "", fmt.Errorf("remote origin is not a GitHub repository URL: %s", raw)
}

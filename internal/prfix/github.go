package prfix

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

type GitHubClient struct{ runner Runner }

func NewGitHubClient(runner Runner) *GitHubClient { return &GitHubClient{runner: runner} }

func (client *GitHubClient) CurrentRepository(ctx context.Context, directory string) (Repository, error) {
	output, err := client.runner.Run(ctx, directory, "git", "remote", "get-url", "origin")
	if err != nil {
		return Repository{}, fmt.Errorf("read Git origin: %w", err)
	}
	repository, err := parseRemote(strings.TrimSpace(string(output)))
	if err != nil {
		return Repository{}, err
	}
	return repository, nil
}

func parseRemote(raw string) (Repository, error) {
	path := raw
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil || !strings.EqualFold(parsed.Hostname(), "github.com") {
			return Repository{}, fmt.Errorf("origin is not a GitHub repository URL: %s", raw)
		}
		path = parsed.Path
	} else if colon := strings.Index(raw, ":"); colon > 0 && !filepath.IsAbs(raw) {
		host := raw[:colon]
		if at := strings.LastIndex(host, "@"); at >= 0 {
			host = host[at+1:]
		}
		if !strings.EqualFold(host, "github.com") {
			return Repository{}, fmt.Errorf("origin is not a GitHub repository URL: %s", raw)
		}
		path = raw[colon+1:]
	}
	parts := strings.Split(strings.Trim(strings.TrimSuffix(path, ".git"), "/"), "/")
	if len(parts) < 2 {
		return Repository{}, fmt.Errorf("origin does not end in owner/repository: %s", raw)
	}
	repository := Repository{Owner: parts[len(parts)-2], Name: parts[len(parts)-1]}
	if err := validateRepository(repository); err != nil {
		return Repository{}, err
	}
	return repository, nil
}

func (client *GitHubClient) ListOpenPullRequests(
	ctx context.Context,
	directory string,
	repository Repository,
) ([]OpenPullRequest, error) {
	output, err := client.runner.Run(ctx, directory, "gh", "pr", "list", "--repo", repository.Slug(),
		"--state", "open", "--limit", "1000", "--json",
		"number,title,url,headRefName,baseRefName,isDraft,reviewDecision")
	if err != nil {
		return nil, fmt.Errorf("list open pull requests: %w", err)
	}
	var pullRequests []OpenPullRequest
	if err := json.Unmarshal(output, &pullRequests); err != nil {
		return nil, fmt.Errorf("decode open pull requests: %w", err)
	}
	return pullRequests, nil
}

func (client *GitHubClient) PullRequest(
	ctx context.Context,
	directory string,
	target Target,
) (PullRequest, error) {
	output, err := client.runner.Run(ctx, directory, "gh", "pr", "view", fmt.Sprint(target.Number),
		"--repo", target.Slug(), "--json",
		"number,title,url,state,headRefName,headRefOid,baseRefName,isCrossRepository")
	if err != nil {
		return PullRequest{}, fmt.Errorf("inspect pull request: %w", err)
	}
	var pullRequest PullRequest
	if err := json.Unmarshal(output, &pullRequest); err != nil {
		return PullRequest{}, fmt.Errorf("decode pull request: %w", err)
	}
	if !strings.EqualFold(pullRequest.State, "OPEN") {
		return PullRequest{}, fmt.Errorf("pull request %s#%d is %s, not open", target.Slug(), target.Number, pullRequest.State)
	}
	if pullRequest.IsCrossRepository {
		return PullRequest{}, fmt.Errorf("pull request %s#%d is from a fork", target.Slug(), target.Number)
	}
	if pullRequest.HeadRefName == "" || pullRequest.HeadRefOID == "" {
		return PullRequest{}, fmt.Errorf("pull request %s#%d has no writable head identity", target.Slug(), target.Number)
	}
	return pullRequest, nil
}

type graphQLErrorEnvelope struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func validateGraphQL(body []byte) error {
	var envelope graphQLErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode GitHub GraphQL response: %w", err)
	}
	if len(envelope.Errors) == 0 {
		return nil
	}
	messages := make([]string, 0, len(envelope.Errors))
	for _, item := range envelope.Errors {
		messages = append(messages, item.Message)
	}
	return fmt.Errorf("GitHub GraphQL error: %s", strings.Join(messages, "; "))
}

func classifyHTTPError(err error, metadata responseMetadata, resetAt string) error {
	if err == nil {
		return nil
	}
	status := metadata.status
	if status != 403 && status != 429 {
		return err
	}
	return &HTTPError{Status: status, RetryAfter: metadata.retryAfter, ResetAt: resetAt, Cause: err}
}

var _ GitHub = (*GitHubClient)(nil)

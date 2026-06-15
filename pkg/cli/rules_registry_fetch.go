package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

type githubContentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
	Path        string `json:"path"`
}

type githubCommitResponse struct {
	SHA string `json:"sha"`
}

func fetchGitHubRulesetRegistry(ctx context.Context) ([]registryRuleset, error) {
	sourceCommit, err := fetchGitHubRegistryCommit(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rulesetRegistryAPIURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ruleset registry from GitHub: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch ruleset registry from GitHub: %s", resp.Status)
	}

	var entries []githubContentEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub ruleset registry: %w", err)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var rulesets []registryRuleset
	for _, entry := range entries {
		if entry.Type != "file" || !strings.HasSuffix(entry.Name, ".md") {
			continue
		}
		if strings.TrimSpace(entry.DownloadURL) == "" {
			return nil, fmt.Errorf("registry ruleset %s has no download URL", entry.Name)
		}
		ruleset, err := fetchGitHubRegistryRuleset(ctx, entry, sourceCommit)
		if err != nil {
			return nil, err
		}
		rulesets = append(rulesets, ruleset)
	}
	return projectRulesetRegistry(rulesets), nil
}

func fetchGitHubRegistryCommit(ctx context.Context) (string, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/commits/%s",
		rulesetRegistryOwner,
		rulesetRegistryRepo,
		rulesetRegistryBranch,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch ruleset registry source commit: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch ruleset registry source commit: %s", resp.Status)
	}
	var payload githubCommitResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("failed to decode ruleset registry source commit: %w", err)
	}
	if strings.TrimSpace(payload.SHA) == "" {
		return "", fmt.Errorf("ruleset registry source commit response had no sha")
	}
	return payload.SHA, nil
}

func fetchGitHubRegistryRuleset(ctx context.Context, entry githubContentEntry, sourceCommit string) (registryRuleset, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.DownloadURL, nil)
	if err != nil {
		return registryRuleset{}, err
	}
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return registryRuleset{}, fmt.Errorf("failed to fetch registry ruleset %s: %w", entry.Name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return registryRuleset{}, fmt.Errorf("failed to fetch registry ruleset %s: %s", entry.Name, resp.Status)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return registryRuleset{}, fmt.Errorf("failed to read registry ruleset %s: %w", entry.Name, err)
	}

	slug := strings.TrimSuffix(entry.Name, ".md")
	parsed := parseRuleset(string(content), entry.Name)
	if issues := validateRulesetDocument(parsed, slug); len(issues) > 0 {
		return registryRuleset{}, fmt.Errorf("registry ruleset %s is invalid: %s", entry.Name, strings.Join(issues, "; "))
	}
	sourcePath := strings.TrimSpace(entry.Path)
	if sourcePath == "" {
		sourcePath = rulesetTarget(parsed.Metadata.Slug)
	}
	normalizedHash, err := normalizedRulesetContentHash(string(content), parsed.Metadata.Status)
	if err != nil {
		return registryRuleset{}, fmt.Errorf("failed to hash registry ruleset %s: %w", entry.Name, err)
	}
	return registryRuleset{
		Slug:           parsed.Metadata.Slug,
		Content:        string(content),
		Metadata:       parsed.Metadata,
		SourceRepo:     rulesetRegistryRepoFullName(),
		SourceBranch:   rulesetRegistryBranch,
		SourceCommit:   sourceCommit,
		SourcePath:     sourcePath,
		NormalizedHash: normalizedHash,
	}, nil
}

func fetchGitHubRegistryContent(ctx context.Context, sourceRepo, sourceCommit, sourcePath string) (string, error) {
	if strings.TrimSpace(sourceRepo) == "" || strings.TrimSpace(sourceCommit) == "" || strings.TrimSpace(sourcePath) == "" {
		return "", fmt.Errorf("source repo, commit, and path are required")
	}
	url := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s",
		strings.TrimSpace(sourceRepo),
		strings.TrimSpace(sourceCommit),
		strings.TrimLeft(strings.TrimSpace(sourcePath), "/"),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch registry artifact base content: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch registry artifact base content: %s", resp.Status)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read registry artifact base content: %w", err)
	}
	return string(content), nil
}

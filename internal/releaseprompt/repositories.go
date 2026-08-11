package releaseprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var skippedRootDirectories = map[string]bool{
	".git": true, ".cache": true, "node_modules": true, "third_party": true, "vendor": true,
}

type ghRepository struct {
	NameWithOwner    string `json:"nameWithOwner"`
	DefaultBranchRef struct {
		Name string `json:"name"`
	} `json:"defaultBranchRef"`
}

func discoverRepositories(ctx context.Context, input Input, runner Runner) ([]Repository, error) {
	paths, err := repositoryPaths(ctx, input, runner)
	if err != nil {
		return nil, err
	}
	_, ghAvailableErr := runner.LookPath("gh")
	ghAvailable := ghAvailableErr == nil
	ghCache := map[string]ghRepository{}
	repositories := make([]Repository, 0, len(paths))
	for _, path := range paths {
		repository, err := inspectRepository(ctx, path, runner, ghAvailable, ghCache)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, repository)
	}
	return repositories, nil
}

func repositoryPaths(ctx context.Context, input Input, runner Runner) ([]string, error) {
	if len(input.Repositories) > 0 && strings.TrimSpace(input.Root) != "" {
		return nil, fmt.Errorf("--repos and --root are mutually exclusive")
	}
	var candidates []string
	if len(input.Repositories) > 0 {
		candidates = append(candidates, input.Repositories...)
	} else if strings.TrimSpace(input.Root) != "" {
		root, err := canonicalDirectory(input.Root)
		if err != nil {
			return nil, fmt.Errorf("resolve repository root: %w", err)
		}
		candidates = append(candidates, root)
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, fmt.Errorf("read repository root %s: %w", root, err)
		}
		for _, entry := range entries {
			if skippedRootDirectories[entry.Name()] || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			path := filepath.Join(root, entry.Name())
			info, err := os.Stat(path)
			if err == nil && info.IsDir() {
				candidates = append(candidates, path)
			}
		}
		sort.Strings(candidates[1:])
	} else {
		return nil, fmt.Errorf("repository scope is required; use --repos or --root")
	}

	seen := map[string]bool{}
	var paths []string
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return nil, fmt.Errorf("repository path cannot be empty")
		}
		directory, err := canonicalDirectory(candidate)
		if err != nil {
			if len(input.Repositories) > 0 {
				return nil, fmt.Errorf("resolve repository %q: %w", candidate, err)
			}
			continue
		}
		top, err := gitText(ctx, runner, directory, "rev-parse", "--show-toplevel")
		if err != nil {
			if len(input.Repositories) > 0 {
				return nil, fmt.Errorf("%s is not a Git repository", directory)
			}
			continue
		}
		top, err = canonicalDirectory(top)
		if err != nil {
			return nil, fmt.Errorf("resolve Git top-level for %s: %w", directory, err)
		}
		if !seen[top] {
			seen[top] = true
			paths = append(paths, top)
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("repository scope did not resolve to any Git repositories")
	}
	return paths, nil
}

func inspectRepository(ctx context.Context, path string, runner Runner, ghAvailable bool, cache map[string]ghRepository) (Repository, error) {
	remote, _ := gitText(ctx, runner, path, "remote", "get-url", "origin")
	identity := githubIdentity(remote)
	branch, _ := gitText(ctx, runner, path, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	branch = cleanBranch(strings.TrimPrefix(branch, "origin/"))
	if ghAvailable && (identity == "" || branch == "") {
		key := identity
		if key == "" {
			key = path
		}
		metadata, ok := cache[key]
		if !ok {
			args := []string{"repo", "view"}
			if identity != "" {
				args = append(args, identity)
			}
			args = append(args, "--json", "nameWithOwner,defaultBranchRef")
			lookupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			output, err := runner.Run(lookupCtx, path, "gh", args...)
			cancel()
			if err == nil && json.Unmarshal(output, &metadata) == nil {
				cache[key] = metadata
			}
		}
		if identity == "" {
			identity = cleanGitHubPath(metadata.NameWithOwner)
		}
		if branch == "" {
			branch = cleanBranch(metadata.DefaultBranchRef.Name)
		}
	}
	name := filepath.Base(path)
	if fields := strings.Split(identity, "/"); len(fields) == 2 && fields[1] != "" {
		name = fields[1]
	}
	return Repository{
		Path: path, Name: name, GitHub: identity, DefaultBranch: branch,
		KitManaged:             kitManaged(path),
		ReleaseWorkflowPresent: fileExists(filepath.Join(path, "docs", "references", "workflows", "release-orchestration.md")),
		VerificationHint:       firstScriptHint(path, verificationCandidates),
		IntegrationSuiteHint:   firstScriptHint(path, integrationCandidates),
	}, nil
}

func canonicalDirectory(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", resolved)
	}
	if strings.ContainsAny(resolved, "\r\n\x00") {
		return "", fmt.Errorf("repository path contains unsupported control characters")
	}
	return filepath.Clean(resolved), nil
}

func gitText(ctx context.Context, runner Runner, dir string, args ...string) (string, error) {
	output, err := runner.Run(ctx, dir, "git", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func githubIdentity(remote string) string {
	remote = strings.TrimSpace(remote)
	if strings.HasPrefix(remote, "git@github.com:") {
		return cleanGitHubPath(strings.TrimPrefix(remote, "git@github.com:"))
	}
	parsed, err := url.Parse(remote)
	if err != nil || !strings.EqualFold(parsed.Hostname(), "github.com") {
		return ""
	}
	return cleanGitHubPath(parsed.Path)
}

func cleanGitHubPath(path string) string {
	path, _, _ = strings.Cut(path, "?")
	path, _, _ = strings.Cut(path, "#")
	path = strings.Trim(strings.TrimSuffix(path, ".git"), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ContainsAny(path, "\r\n\x00`") {
		return ""
	}
	return parts[0] + "/" + parts[1]
}

func cleanBranch(branch string) string {
	branch = strings.TrimSpace(branch)
	if strings.ContainsAny(branch, "\r\n\x00` ") {
		return ""
	}
	return branch
}

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func resolvePromptWorktreeRoot(start string) (string, error) {
	if strings.TrimSpace(start) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		start = cwd
	}
	output, err := repairContextCommandOutput(
		context.Background(),
		start,
		"git",
		"rev-parse",
		"--show-toplevel",
	)
	if err == nil {
		if root := strings.TrimSpace(string(output)); root != "" {
			return filepath.Clean(root), nil
		}
	}
	absolute, absErr := filepath.Abs(start)
	if absErr != nil {
		return "", fmt.Errorf("resolve working directory: %w", absErr)
	}
	return filepath.Clean(absolute), nil
}

func repairGitText(ctx context.Context, dir string, args ...string) (string, error) {
	output, err := repairContextCommandOutput(ctx, dir, "git", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func runRepairContextCommand(
	ctx context.Context,
	dir string,
	name string,
	args ...string,
) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := execCommandContext(ctx, name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err == nil {
		return output, nil
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return nil, err
	}
	return nil, fmt.Errorf("%w: %s", err, detail)
}

func inferOpenPRForBranch(
	ctx context.Context,
	cwd string,
	repository string,
	branch string,
) (string, bool, error) {
	output, err := repairContextCommandOutput(
		ctx,
		cwd,
		"gh",
		"pr",
		"list",
		"--repo",
		repository,
		"--state",
		"open",
		"--head",
		branch,
		"--json",
		"number,url",
	)
	if err != nil {
		return "", false, fmt.Errorf("find open PR for %s branch %s: %w", repository, branch, err)
	}
	var pullRequests []struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(output, &pullRequests); err != nil {
		return "", false, fmt.Errorf("parse open PRs for %s branch %s: %w", repository, branch, err)
	}
	switch len(pullRequests) {
	case 0:
		return "", false, nil
	case 1:
		if strings.TrimSpace(pullRequests[0].URL) != "" {
			return pullRequests[0].URL, true, nil
		}
		return repository + "#" + strconv.Itoa(pullRequests[0].Number), true, nil
	default:
		return "", false, fmt.Errorf(
			"multiple open pull requests use %s branch %s; target one explicitly",
			repository,
			branch,
		)
	}
}

func resolveRepositoryDefaultBranch(
	ctx context.Context,
	cwd string,
	repository string,
) (string, error) {
	output, err := repairContextCommandOutput(
		ctx,
		cwd,
		"gh",
		"repo",
		"view",
		repository,
		"--json",
		"defaultBranchRef",
		"-q",
		".defaultBranchRef.name",
	)
	if err != nil {
		return "", fmt.Errorf("resolve default branch for %s: %w", repository, err)
	}
	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "", fmt.Errorf("repository %s did not report a default branch", repository)
	}
	return branch, nil
}

func prepareCIRepairContext(
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	cwd string,
	diagnosis ciDiagnosis,
) (*repairContext, error) {
	target := diagnosis.Target
	if target.PRNumber > 0 {
		return resolvePRRepairContext(
			ctx,
			in,
			out,
			cwd,
			target.Repository+"#"+strconv.Itoa(target.PRNumber),
		)
	}
	branch := strings.TrimSpace(target.Branch)
	if branch == "" {
		return nil, nil
	}
	if prRef, found, err := inferOpenPRForBranch(ctx, cwd, target.Repository, branch); err != nil {
		return nil, err
	} else if found {
		return resolvePRRepairContext(ctx, in, out, cwd, prRef)
	}
	defaultBranch, err := resolveRepositoryDefaultBranch(ctx, cwd, target.Repository)
	if err != nil {
		return nil, err
	}
	if branch == defaultBranch {
		return nil, fmt.Errorf(
			"CI target %s is the protected default branch for %s and has no open PR; Kit cannot infer a writable repair lane",
			branch,
			target.Repository,
		)
	}
	return prepareBranchRepairContext(
		ctx,
		in,
		out,
		cwd,
		target.Repository,
		branch,
		target.HeadSHA,
	)
}

package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/worktree"
)

type repairChangeDisposition string

const (
	repairChangesNone    repairChangeDisposition = "none"
	repairChangesInclude repairChangeDisposition = "include"
	repairChangesExclude repairChangeDisposition = "exclude"
)

type repairContext struct {
	Repository        string
	PRNumber          int
	PRURL             string
	HeadBranch        string
	ExpectedHeadOID   string
	LocalHeadOID      string
	WorktreePath      string
	WorktreeCreated   bool
	ExistingChanges   repairChangeDisposition
	DirtyStatus       string
	PushTarget        string
	TargetDescription string
}

var (
	preparePullRequestWorktree = func(
		ctx context.Context,
		cwd string,
		number int,
	) (worktree.PullRequestRepair, error) {
		return worktree.NewApp(io.Discard, io.Discard).
			PreparePullRequestRepair(ctx, cwd, number, true)
	}
	prepareBranchWorktree = func(
		ctx context.Context,
		cwd string,
		branch string,
	) (worktree.PreparedWorktree, error) {
		return worktree.NewApp(io.Discard, io.Discard).
			PrepareBranch(ctx, cwd, branch, true)
	}
	repairContextCommandOutput = runRepairContextCommand
	resolvePRRepairContext     = preparePRRepairContext
)

func preparePRRepairContext(
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	cwd string,
	prRef string,
) (*repairContext, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	target, err := resolveDispatchPRTarget(prRef)
	if err != nil {
		return nil, err
	}
	repository := target.Owner + "/" + target.Repo
	if err := requireRepairRepository(cwd, repository); err != nil {
		return nil, err
	}

	prepared, err := preparePullRequestWorktree(ctx, cwd, target.Number)
	if err != nil {
		return nil, fmt.Errorf("prepare writable worktree for %s#%d: %w", repository, target.Number, err)
	}
	if !strings.EqualFold(prepared.Repository, repository) {
		return nil, fmt.Errorf(
			"prepared worktree belongs to %s, not requested repository %s",
			prepared.Repository,
			repository,
		)
	}
	if strings.TrimSpace(prepared.HeadRefOID) == "" {
		return nil, fmt.Errorf("PR %s#%d did not report a head SHA", repository, target.Number)
	}

	prURL := strings.TrimSpace(prepared.URL)
	if prURL == "" {
		prURL = fmt.Sprintf("https://github.com/%s/pull/%d", repository, target.Number)
	}
	return inspectRepairContext(in, out, repairContext{
		Repository:        repository,
		PRNumber:          target.Number,
		PRURL:             prURL,
		HeadBranch:        prepared.Branch,
		ExpectedHeadOID:   prepared.HeadRefOID,
		WorktreePath:      prepared.Path,
		WorktreeCreated:   prepared.Created,
		PushTarget:        "origin/" + prepared.Branch,
		TargetDescription: prURL,
	})
}

func prepareBranchRepairContext(
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	cwd string,
	repository string,
	branch string,
	expectedHeadOID string,
) (*repairContext, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireRepairRepository(cwd, repository); err != nil {
		return nil, err
	}
	prepared, err := prepareBranchWorktree(ctx, cwd, branch)
	if err != nil {
		return nil, fmt.Errorf("prepare writable worktree for %s branch %s: %w", repository, branch, err)
	}
	return inspectRepairContext(in, out, repairContext{
		Repository:        repository,
		HeadBranch:        branch,
		ExpectedHeadOID:   strings.TrimSpace(expectedHeadOID),
		WorktreePath:      prepared.Path,
		WorktreeCreated:   prepared.Created,
		PushTarget:        "origin/" + branch,
		TargetDescription: repository + " branch " + branch,
	})
}

func inspectRepairContext(
	in io.Reader,
	out io.Writer,
	repair repairContext,
) (*repairContext, error) {
	branch, err := repairGitText(repair.WorktreePath, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("verify repair worktree branch: %w", err)
	}
	if branch != repair.HeadBranch {
		return nil, fmt.Errorf(
			"repair worktree %s is on branch %q, expected %q",
			repair.WorktreePath,
			branch,
			repair.HeadBranch,
		)
	}
	repair.LocalHeadOID, err = repairGitText(repair.WorktreePath, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("read repair worktree HEAD: %w", err)
	}
	repair.DirtyStatus, err = repairGitText(
		repair.WorktreePath,
		"status",
		"--porcelain=v1",
		"--untracked-files=all",
	)
	if err != nil {
		return nil, fmt.Errorf("inspect repair worktree status: %w", err)
	}
	action := "Reusing"
	if repair.WorktreeCreated {
		action = "Created"
	}
	if _, err := fmt.Fprintf(
		out,
		"%s repair worktree %s for %s\n",
		action,
		repair.WorktreePath,
		repair.TargetDescription,
	); err != nil {
		return nil, err
	}
	repair.ExistingChanges = repairChangesNone
	if repair.DirtyStatus != "" {
		include, err := confirmRepairChanges(in, out, repair)
		if err != nil {
			return nil, err
		}
		repair.ExistingChanges = repairChangesExclude
		if include {
			repair.ExistingChanges = repairChangesInclude
		}
	}
	return &repair, nil
}

func confirmRepairChanges(in io.Reader, out io.Writer, repair repairContext) (bool, error) {
	target := "target branch repair"
	if repair.PRURL != "" {
		target = "existing pull request repair"
	}
	if _, err := fmt.Fprintf(
		out,
		"Existing changes in %s for %s:\n%s\nInclude these changes in the %s? [y/N]: ",
		repair.WorktreePath,
		repair.TargetDescription,
		repair.DirtyStatus,
		target,
	); err != nil {
		return false, err
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err == io.EOF && line == "" {
		return false, fmt.Errorf(
			"dirty repair worktree requires an explicit y or n response",
		)
	}
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read repair worktree confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	case "", "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf(
			"invalid repair worktree confirmation %q; enter y or n",
			strings.TrimSpace(line),
		)
	}
}

func requireRepairRepository(cwd string, expected string) error {
	output, err := repairContextCommandOutput(cwd, "git", "remote", "get-url", "origin")
	if err != nil {
		return fmt.Errorf("resolve repair repository from origin: %w", err)
	}
	owner, repo, err := parseGitHubRemoteURL(strings.TrimSpace(string(output)))
	if err != nil {
		return err
	}
	actual := owner + "/" + repo
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf(
			"requested repair target %s does not belong to current clone %s; run Kit from a checkout of the target repository",
			expected,
			actual,
		)
	}
	return nil
}

func resolvePromptWorktreeRoot(start string) (string, error) {
	if strings.TrimSpace(start) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		start = cwd
	}
	output, err := repairContextCommandOutput(start, "git", "rev-parse", "--show-toplevel")
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

func repairGitText(dir string, args ...string) (string, error) {
	output, err := repairContextCommandOutput(dir, "git", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func runRepairContextCommand(dir string, name string, args ...string) ([]byte, error) {
	cmd := execCommand(name, args...)
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
	cwd string,
	repository string,
	branch string,
) (string, bool, error) {
	output, err := repairContextCommandOutput(
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

func resolveRepositoryDefaultBranch(cwd string, repository string) (string, error) {
	output, err := repairContextCommandOutput(
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
	if prRef, found, err := inferOpenPRForBranch(cwd, target.Repository, branch); err != nil {
		return nil, err
	} else if found {
		return resolvePRRepairContext(ctx, in, out, cwd, prRef)
	}
	defaultBranch, err := resolveRepositoryDefaultBranch(cwd, target.Repository)
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

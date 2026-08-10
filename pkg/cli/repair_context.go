package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/jamesonstone/kit/internal/worktreeprep"
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
	) (worktreeprep.PullRequest, error) {
		return worktreeprep.New().PreparePullRequest(ctx, cwd, number, true)
	}
	prepareBranchWorktree = func(
		ctx context.Context,
		cwd string,
		branch string,
	) (worktreeprep.Prepared, error) {
		return worktreeprep.New().PrepareBranch(ctx, cwd, branch, true)
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
	if err := requireRepairRepository(ctx, cwd, repository); err != nil {
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
	return inspectRepairContext(ctx, in, out, repairContext{
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
	if err := requireRepairRepository(ctx, cwd, repository); err != nil {
		return nil, err
	}
	prepared, err := prepareBranchWorktree(ctx, cwd, branch)
	if err != nil {
		return nil, fmt.Errorf("prepare writable worktree for %s branch %s: %w", repository, branch, err)
	}
	return inspectRepairContext(ctx, in, out, repairContext{
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
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	repair repairContext,
) (*repairContext, error) {
	branch, err := repairGitText(ctx, repair.WorktreePath, "symbolic-ref", "--quiet", "--short", "HEAD")
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
	repair.LocalHeadOID, err = repairGitText(ctx, repair.WorktreePath, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("read repair worktree HEAD: %w", err)
	}
	repair.DirtyStatus, err = repairGitText(
		ctx,
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

func requireRepairRepository(ctx context.Context, cwd string, expected string) error {
	output, err := repairContextCommandOutput(ctx, cwd, "git", "remote", "get-url", "origin")
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

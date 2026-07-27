package worktree

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func (a *App) liveOriginDefault(
	ctx context.Context,
	cwd string,
) (string, string, error) {
	output, err := a.git(ctx, cwd, "ls-remote", "--symref", "origin", "HEAD")
	if err != nil {
		return "", "", fmt.Errorf("discover live origin default branch: %w", err)
	}
	var branch string
	var oid string
	for _, line := range strings.Split(string(output), "\n") {
		switch {
		case strings.HasPrefix(line, "ref: refs/heads/") &&
			strings.HasSuffix(line, "\tHEAD"):
			branch = strings.TrimSuffix(
				strings.TrimPrefix(line, "ref: refs/heads/"),
				"\tHEAD",
			)
		case strings.HasSuffix(line, "\tHEAD") && !strings.HasPrefix(line, "ref: "):
			oid = strings.TrimSuffix(line, "\tHEAD")
		}
	}
	if branch == "" || oid == "" {
		return "", "", fmt.Errorf("origin did not advertise a symbolic default branch and OID")
	}
	if _, err := validateLane(branch); err != nil {
		return "", "", fmt.Errorf("origin advertised unsafe default branch %q: %w", branch, err)
	}
	return branch, oid, nil
}

func (a *App) reconcileDefaultBranch(
	ctx context.Context,
	repo repository,
	entries []worktreeEntry,
	branch string,
	remoteOID string,
	dryRun bool,
	report *SyncReport,
) SyncDefaultDecision {
	decision := SyncDefaultDecision{
		Branch:    branch,
		RemoteOID: remoteOID,
		Action:    "preserved",
	}
	localOID, err := a.gitText(
		ctx,
		repo.top,
		"rev-parse",
		"--verify",
		"refs/heads/"+branch,
	)
	if err != nil {
		decision.State = "missing"
		decision.Detail = "local default branch does not exist"
		return decision
	}
	decision.LocalOID = localOID
	for _, entry := range entries {
		if entry.branch == branch {
			decision.Path = entry.path
			break
		}
	}
	if localOID == remoteOID {
		decision.State = "current"
		decision.Action = "none"
		return decision
	}

	if decision.Path != "" {
		dirty, statusErr := a.status(ctx, decision.Path, false)
		if statusErr != nil {
			decision.State = "unavailable"
			decision.Action = "failed"
			decision.Detail = statusErr.Error()
			report.addFailure("inspect-default-status", decision.Path, statusErr)
			return decision
		}
		if dirty != "" {
			decision.State = "dirty"
			decision.Detail = dirty
			return decision
		}
	}

	if _, objectErr := a.git(
		ctx,
		repo.top,
		"cat-file",
		"-e",
		remoteOID+"^{commit}",
	); objectErr != nil {
		decision.State = "remote-object-unavailable"
		decision.Detail = "strict dry-run did not fetch the live origin commit"
		return decision
	}
	localBehind, localBehindErr := a.isAncestor(ctx, repo.top, localOID, remoteOID)
	remoteBehind, remoteBehindErr := a.isAncestor(ctx, repo.top, remoteOID, localOID)
	if localBehindErr != nil || remoteBehindErr != nil {
		ancestryErr := errors.Join(localBehindErr, remoteBehindErr)
		decision.State = "unavailable"
		decision.Action = "failed"
		decision.Detail = ancestryErr.Error()
		report.addFailure("classify-default", decision.Path, ancestryErr)
		return decision
	}
	switch {
	case localBehind && !remoteBehind:
		decision.State = "behind"
		if dryRun {
			decision.Action = "would-fast-forward"
			return decision
		}
		if err := a.fastForwardDefault(ctx, repo, branch, remoteOID, localOID, decision.Path); err != nil {
			decision.Action = "failed"
			decision.Detail = err.Error()
			operation := "advance-default-ref"
			if decision.Path != "" {
				operation = "fast-forward-default"
			}
			report.addFailure(operation, decision.Path, err)
			return decision
		}
		reachedOID, err := a.gitText(
			ctx,
			repo.top,
			"rev-parse",
			"--verify",
			"refs/heads/"+branch,
		)
		if err != nil {
			verificationErr := fmt.Errorf("verify fast-forwarded default branch: %w", err)
			decision.Action = "failed"
			decision.Detail = verificationErr.Error()
			report.addFailure("verify-default-ref", decision.Path, verificationErr)
			return decision
		}
		if reachedOID != remoteOID {
			err := fmt.Errorf(
				"fast-forwarded default branch to %s, expected %s",
				reachedOID,
				remoteOID,
			)
			decision.Action = "failed"
			decision.Detail = err.Error()
			report.addFailure("verify-default-ref", decision.Path, err)
			decision.LocalOID = reachedOID
			return decision
		}
		decision.Action = "fast-forwarded"
		decision.LocalOID = reachedOID
	case remoteBehind && !localBehind:
		decision.State = "ahead"
	case !localBehind && !remoteBehind:
		decision.State = "diverged"
	default:
		decision.State = "unavailable"
		decision.Action = "failed"
		decision.Detail = "could not classify default branch ancestry"
		report.addFailure(
			"classify-default",
			decision.Path,
			errors.New(decision.Detail),
		)
	}
	return decision
}

func (a *App) fastForwardDefault(
	ctx context.Context,
	repo repository,
	branch string,
	remoteOID string,
	localOID string,
	worktreePath string,
) error {
	if worktreePath != "" {
		_, err := a.git(
			ctx,
			worktreePath,
			"merge",
			"--ff-only",
			remoteOID,
		)
		return err
	}
	_, err := a.git(
		ctx,
		repo.top,
		"update-ref",
		"refs/heads/"+branch,
		remoteOID,
		localOID,
	)
	return err
}

func (a *App) isAncestor(
	ctx context.Context,
	cwd string,
	ancestor string,
	descendant string,
) (bool, error) {
	output, err := a.run(
		ctx,
		cwd,
		"git",
		"merge-base",
		"--is-ancestor",
		ancestor,
		descendant,
	)
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return false, fmt.Errorf(
			"git merge-base --is-ancestor %s %s: %w",
			ancestor,
			descendant,
			err,
		)
	}
	return false, fmt.Errorf(
		"git merge-base --is-ancestor %s %s: %w\n%s",
		ancestor,
		descendant,
		err,
		detail,
	)
}

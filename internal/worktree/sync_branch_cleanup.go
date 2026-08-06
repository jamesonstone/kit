package worktree

import (
	"context"
	"fmt"
	"strings"
)

func (a *App) prepareMergedLocalBranchDeletion(
	ctx context.Context,
	repo repository,
	branch string,
	expectedOID string,
) (string, error) {
	mergeOutput, err := a.git(
		ctx,
		repo.top,
		"config",
		"--get-all",
		"branch."+branch+".merge",
	)
	if err != nil {
		return "", fmt.Errorf("resolve local branch merge ref: %w", err)
	}
	if mergeRef := strings.TrimSpace(string(mergeOutput)); mergeRef != "refs/heads/"+branch {
		return "", fmt.Errorf(
			"local branch merge ref %q is not the exact head ref refs/heads/%s",
			mergeRef,
			branch,
		)
	}
	missingFetch := "__kit_wt_sync_missing_" + expectedOID + "__"
	configuredFetch, err := a.git(
		ctx,
		repo.top,
		"config",
		"--default",
		missingFetch,
		"--get",
		"remote.kit-wt-sync-proof.fetch",
	)
	if err != nil {
		return "", fmt.Errorf("inspect temporary proof remote: %w", err)
	}
	if fetchRefspec := strings.TrimSpace(string(configuredFetch)); fetchRefspec != missingFetch {
		return "", fmt.Errorf(
			"temporary proof remote name collides with configured fetch %q",
			fetchRefspec,
		)
	}

	proofRef := "refs/remotes/kit-wt-sync-proof/" + branch
	zeroOID := strings.Repeat("0", len(expectedOID))
	_, err = a.git(
		ctx,
		repo.top,
		"update-ref",
		"--no-deref",
		proofRef,
		expectedOID,
		zeroOID,
	)
	if err != nil {
		return "", fmt.Errorf("create exact branch-deletion proof: %w", err)
	}
	return proofRef, nil
}

func (a *App) deleteMergedLocalBranch(
	ctx context.Context,
	repo repository,
	branch string,
	expectedOID string,
	proofRef string,
) error {
	branchRef := "refs/heads/" + branch
	output, err := a.git(ctx, repo.top, "rev-parse", "--verify", branchRef)
	if err != nil {
		return fmt.Errorf("verify exact local branch: %w", err)
	}
	if actualOID := strings.TrimSpace(string(output)); actualOID != expectedOID {
		return fmt.Errorf(
			"local branch moved from proven PR head %s to %s",
			expectedOID,
			detailOr(actualOID, "missing"),
		)
	}
	_, err = a.git(
		ctx,
		repo.top,
		"-c",
		"branch."+branch+".remote=kit-wt-sync-proof",
		"-c",
		"remote.kit-wt-sync-proof.fetch=+refs/heads/*:refs/remotes/kit-wt-sync-proof/*",
		"branch",
		"-d",
		"--",
		branch,
	)
	return err
}

func (a *App) cleanupMergedLocalBranchDeletion(
	ctx context.Context,
	repo repository,
	proofRef string,
	expectedOID string,
) error {
	_, err := a.git(
		ctx,
		repo.top,
		"update-ref",
		"--no-deref",
		"-d",
		proofRef,
		expectedOID,
	)
	if err != nil {
		return fmt.Errorf("remove exact branch-deletion proof: %w", err)
	}
	return nil
}

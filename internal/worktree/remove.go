package worktree

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type managedEnvironmentLink struct {
	path   string
	target string
}

type worktreeRemoval struct {
	entry           worktreeEntry
	environmentLink *managedEnvironmentLink
}

type worktreeDirtyError struct {
	path   string
	status string
}

func (err worktreeDirtyError) Error() string {
	return fmt.Sprintf(
		"%s contains tracked, untracked, or ignored material; refusing removal:\n%s",
		err.path,
		err.status,
	)
}

func (a *App) remove(ctx context.Context, cwd, target string) error {
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	path := target
	if !filepath.IsAbs(path) {
		path, err = canonicalLanePath(repo, target)
		if err != nil {
			return err
		}
	}
	path = filepath.Clean(path)
	relative, err := filepath.Rel(repo.projectRoot, path)
	if err != nil || relative == "." || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("target must be one exact worktree beneath %s", repo.projectRoot)
	}
	if samePath(path, repo.top) {
		return fmt.Errorf("refusing to remove the current worktree")
	}

	selected, err := a.registeredWorktree(ctx, repo.top, path)
	if err != nil {
		return err
	}
	removal, err := a.inspectWorktreeRemoval(ctx, repo, *selected)
	if err != nil {
		return err
	}
	if err := a.ensurePublished(ctx, *selected); err != nil {
		return err
	}
	if err := a.executeWorktreeRemoval(ctx, repo, removal); err != nil {
		return err
	}
	return a.writef("Removed worktree %s; branch and shared Git state were preserved.\n", selected.path)
}

func (a *App) inspectWorktreeRemoval(
	ctx context.Context,
	repo repository,
	selected worktreeEntry,
) (worktreeRemoval, error) {
	environmentLink, err := inspectManagedEnvironmentLink(repo.primary, selected.path)
	if err != nil {
		return worktreeRemoval{}, err
	}
	dirty, err := a.status(ctx, selected.path, true)
	if err != nil {
		return worktreeRemoval{}, err
	}
	if environmentLink != nil {
		dirty = statusWithoutManagedEnvironmentLink(dirty)
	}
	if dirty != "" {
		return worktreeRemoval{}, worktreeDirtyError{path: selected.path, status: dirty}
	}
	return worktreeRemoval{entry: selected, environmentLink: environmentLink}, nil
}

func (a *App) executeWorktreeRemoval(
	ctx context.Context,
	repo repository,
	removal worktreeRemoval,
) error {
	environmentLink := removal.environmentLink
	if environmentLink != nil {
		if err := os.Remove(environmentLink.path); err != nil {
			return fmt.Errorf("remove managed environment symlink %s: %w", environmentLink.path, err)
		}
	}
	if _, err := a.git(ctx, repo.top, "worktree", "remove", removal.entry.path); err != nil {
		if environmentLink == nil {
			return err
		}
		if restoreErr := os.Symlink(environmentLink.target, environmentLink.path); restoreErr != nil {
			return fmt.Errorf(
				"%w; additionally failed to restore environment symlink %s: %v",
				err,
				environmentLink.path,
				restoreErr,
			)
		}
		return fmt.Errorf("%w; restored environment symlink %s", err, environmentLink.path)
	}
	return nil
}

func (a *App) ensurePublished(ctx context.Context, selected worktreeEntry) error {
	if selected.branch == "" {
		return nil
	}
	upstream, err := a.gitText(
		ctx,
		selected.path,
		"rev-parse",
		"--abbrev-ref",
		"--symbolic-full-name",
		"@{upstream}",
	)
	if err != nil {
		return fmt.Errorf(
			"branch %s has no upstream; refusing removal because published state cannot be proven",
			selected.branch,
		)
	}
	aheadText, err := a.gitText(ctx, selected.path, "rev-list", "--count", upstream+"..HEAD")
	if err != nil {
		return err
	}
	ahead, err := strconv.Atoi(aheadText)
	if err != nil {
		return fmt.Errorf("parse ahead count %q: %w", aheadText, err)
	}
	if ahead != 0 {
		return fmt.Errorf(
			"branch %s is %d commit(s) ahead of %s; refusing removal",
			selected.branch,
			ahead,
			upstream,
		)
	}
	return nil
}

func inspectManagedEnvironmentLink(
	sourceRoot string,
	worktreePath string,
) (*managedEnvironmentLink, error) {
	path := filepath.Join(worktreePath, environmentFileName)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect worktree environment file %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return nil, fmt.Errorf(
			"%s is not a GitWT-managed environment symlink; refusing removal",
			path,
		)
	}
	expectedSource := filepath.Join(sourceRoot, environmentFileName)
	matches, target, err := environmentSymlinkMatches(path, expectedSource)
	if err != nil {
		return nil, err
	}
	if !matches {
		return nil, fmt.Errorf(
			"%s points somewhere other than the expected source %s; refusing removal",
			path,
			expectedSource,
		)
	}
	return &managedEnvironmentLink{path: path, target: target}, nil
}

func statusWithoutManagedEnvironmentLink(status string) string {
	lines := strings.Split(status, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if line == "?? "+environmentFileName || line == "!! "+environmentFileName {
			continue
		}
		if line != "" {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

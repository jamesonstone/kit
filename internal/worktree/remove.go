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
	entry            worktreeEntry
	environmentLinks []managedEnvironmentLink
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
	environmentLinks, err := inspectManagedEnvironmentLinks(repo.primary, selected.path)
	if err != nil {
		return worktreeRemoval{}, err
	}
	dirty, err := a.status(ctx, selected.path, true)
	if err != nil {
		return worktreeRemoval{}, err
	}
	if len(environmentLinks) > 0 {
		dirty = statusWithoutManagedEnvironmentLinks(dirty, environmentLinks)
	}
	if dirty != "" {
		return worktreeRemoval{}, worktreeDirtyError{path: selected.path, status: dirty}
	}
	return worktreeRemoval{entry: selected, environmentLinks: environmentLinks}, nil
}

func (a *App) executeWorktreeRemoval(
	ctx context.Context,
	repo repository,
	removal worktreeRemoval,
) error {
	removedLinks := make([]managedEnvironmentLink, 0, len(removal.environmentLinks))
	for _, environmentLink := range removal.environmentLinks {
		if err := os.Remove(environmentLink.path); err != nil {
			restoreErr := restoreManagedEnvironmentLinks(removedLinks)
			if restoreErr != nil {
				return fmt.Errorf(
					"remove managed environment symlink %s: %w; additionally failed to restore an environment symlink: %v",
					environmentLink.path,
					err,
					restoreErr,
				)
			}
			return fmt.Errorf("remove managed environment symlink %s: %w", environmentLink.path, err)
		}
		removedLinks = append(removedLinks, environmentLink)
	}
	if _, err := a.git(ctx, repo.top, "worktree", "remove", removal.entry.path); err != nil {
		if len(removedLinks) == 0 {
			return err
		}
		if restoreErr := restoreManagedEnvironmentLinks(removedLinks); restoreErr != nil {
			return fmt.Errorf(
				"%w; additionally failed to restore environment symlinks: %v",
				err,
				restoreErr,
			)
		}
		return fmt.Errorf("%w; restored environment symlink(s)", err)
	}
	return nil
}

func restoreManagedEnvironmentLinks(environmentLinks []managedEnvironmentLink) error {
	var restoreErrors []error
	for _, environmentLink := range environmentLinks {
		if err := os.Symlink(environmentLink.target, environmentLink.path); err != nil {
			restoreErrors = append(restoreErrors, fmt.Errorf("%s: %w", environmentLink.path, err))
		}
	}
	return errors.Join(restoreErrors...)
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

func inspectManagedEnvironmentLinks(
	sourceRoot string,
	worktreePath string,
) ([]managedEnvironmentLink, error) {
	environmentLinks := make([]managedEnvironmentLink, 0, len(environmentFileNames))
	for _, name := range environmentFileNames {
		environmentLink, err := inspectManagedEnvironmentLink(sourceRoot, worktreePath, name)
		if err != nil {
			return nil, err
		}
		if environmentLink != nil {
			environmentLinks = append(environmentLinks, *environmentLink)
		}
	}
	return environmentLinks, nil
}

func inspectManagedEnvironmentLink(
	sourceRoot string,
	worktreePath string,
	name string,
) (*managedEnvironmentLink, error) {
	path := filepath.Join(worktreePath, name)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect worktree environment file %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		if name == environmentRCFileName {
			return nil, nil
		}
		return nil, fmt.Errorf(
			"%s is not a GitWT-managed environment symlink; refusing removal",
			path,
		)
	}
	expectedSource := filepath.Join(sourceRoot, name)
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

func statusWithoutManagedEnvironmentLinks(
	status string,
	environmentLinks []managedEnvironmentLink,
) string {
	lines := strings.Split(status, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if isManagedEnvironmentLinkStatus(line, environmentLinks) {
			continue
		}
		if line != "" {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func isManagedEnvironmentLinkStatus(
	line string,
	environmentLinks []managedEnvironmentLink,
) bool {
	for _, environmentLink := range environmentLinks {
		name := filepath.Base(environmentLink.path)
		if line == "?? "+name || line == "!! "+name {
			return true
		}
	}
	return false
}

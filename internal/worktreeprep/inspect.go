package worktreeprep

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// Inspect resolves the current checkout without fetching or changing Git state.
func (preparer *Preparer) Inspect(ctx context.Context, cwd string) (Location, error) {
	output, err := preparer.run(ctx, cwd, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		hasGitMetadata, metadataErr := preparer.hasGitMetadataInAncestors(cwd)
		if metadataErr != nil {
			return Location{}, metadataErr
		}
		detail := strings.TrimSpace(string(output))
		if !hasGitMetadata && strings.Contains(detail, "not a git repository") {
			return Location{}, nil
		}
		return Location{}, fmt.Errorf("resolve current Git worktree: %w: %s", err, detail)
	}
	top := strings.TrimSpace(string(output))
	if top == "" {
		return Location{}, fmt.Errorf("resolve current Git worktree: empty repository root")
	}

	entries, err := preparer.worktrees(ctx, top)
	if err != nil {
		return Location{}, fmt.Errorf("list repository worktrees: %w", err)
	}
	if len(entries) == 0 {
		return Location{}, fmt.Errorf("repository has no primary worktree")
	}

	primary := filepath.Clean(entries[0].path)
	return Location{
		Path:        filepath.Clean(top),
		PrimaryPath: primary,
		InsideGit:   true,
		IsPrimary:   samePath(top, primary),
	}, nil
}

func (preparer *Preparer) hasGitMetadataInAncestors(cwd string) (bool, error) {
	current, err := filepath.Abs(cwd)
	if err != nil {
		return false, fmt.Errorf("resolve worktree inspection path: %w", err)
	}
	for {
		exists, statErr := preparer.pathExists(filepath.Join(current, ".git"))
		if statErr != nil {
			return false, fmt.Errorf("inspect Git metadata at %s: %w", current, statErr)
		}
		if exists {
			return true, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return false, nil
		}
		current = parent
	}
}

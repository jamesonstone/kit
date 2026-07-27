package worktree

import (
	"errors"
	"path/filepath"
	"strings"
)

func baseSyncLaneDecision(
	repo repository,
	entry worktreeEntry,
	defaultBranch string,
) (SyncLaneDecision, bool) {
	decision := SyncLaneDecision{
		Path:    entry.path,
		Branch:  entry.branch,
		HeadOID: entry.head,
		Action:  "preserved",
	}
	switch {
	case samePath(entry.path, repo.primary) && samePath(entry.path, repo.top):
		decision.Reason = "primary-and-current-worktree"
		return decision, false
	case samePath(entry.path, repo.primary):
		decision.Reason = "primary-worktree"
		return decision, false
	case samePath(entry.path, repo.top):
		decision.Reason = "current-worktree"
		return decision, false
	case entry.prunable:
		decision.Reason = "stale-worktree-metadata"
		return decision, false
	case entry.branch == "":
		decision.Reason = "detached-worktree"
		return decision, false
	case entry.branch == defaultBranch:
		decision.Reason = "default-branch-worktree"
		return decision, false
	}
	resolvedRoot, rootErr := resolvedPath(repo.projectRoot)
	resolvedEntry, entryErr := resolvedPath(entry.path)
	if rootErr != nil || entryErr != nil {
		decision.Reason = "non-canonical-worktree"
		return decision, false
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedEntry)
	if err != nil ||
		relative == "." ||
		relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		decision.Reason = "non-canonical-worktree"
		return decision, false
	}
	lane := filepath.ToSlash(relative)
	expected, err := canonicalLanePath(repo, lane)
	if err != nil || !samePath(entry.path, expected) {
		decision.Reason = "non-canonical-worktree"
		return decision, false
	}
	decision.Reason = "awaiting-pull-request-evidence"
	return decision, true
}

func removalRefusalReason(err error) string {
	var dirty worktreeDirtyError
	if errors.As(err, &dirty) {
		return "worktree-dirty"
	}
	message := err.Error()
	if strings.Contains(message, environmentFileName) &&
		(strings.Contains(message, "not a GitWT-managed") ||
			strings.Contains(message, "points somewhere other than")) {
		return "unmanaged-environment-material"
	}
	return "inspection-failed"
}

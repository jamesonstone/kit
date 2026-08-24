package cli

import (
	"strings"
	"testing"
)

const (
	scopedGitCleanFd     = "`git clean -fd` with only those verified paths"
	gitCleanFdLiteral    = "`git clean -fd`"
	pullMergedBranch     = "Then pull the merged default branch"
	confirmMergeFirst    = "Confirm the merge first"
	confirmMergeEvidence = "Confirm merge evidence first"
)

func firstMergeConfirmationIndex(text string) int {
	found := -1
	for _, marker := range []string{confirmMergeFirst, confirmMergeEvidence} {
		idx := strings.Index(text, marker)
		if idx >= 0 && (found < 0 || idx < found) {
			found = idx
		}
	}
	return found
}

func scopedPostMergeCleanupOrderError(text string) string {
	mergeAt := firstMergeConfirmationIndex(text)
	if mergeAt < 0 {
		return "missing merge confirmation instruction"
	}
	scopedAt := strings.Index(text, scopedGitCleanFd)
	if scopedAt < 0 {
		return "missing path-scoped git clean instruction"
	}
	pullAt := strings.Index(text, pullMergedBranch)
	if pullAt < 0 {
		return "missing pull-after-cleanup instruction"
	}
	if mergeAt >= scopedAt || scopedAt >= pullAt {
		return "merge confirmation must precede targeted cleanup, and cleanup must precede pulling"
	}
	for idx := 0; ; {
		next := strings.Index(text[idx:], gitCleanFdLiteral)
		if next < 0 {
			break
		}
		abs := idx + next
		if abs < mergeAt {
			return "cleanup instruction occurs before merge confirmation"
		}
		if !strings.HasPrefix(text[abs:], scopedGitCleanFd) {
			return "unscoped git clean -fd is not allowed"
		}
		idx = abs + len(gitCleanFdLiteral)
	}
	return ""
}

func assertScopedPostMergeCleanupOrder(t *testing.T, text string) {
	t.Helper()
	if msg := scopedPostMergeCleanupOrderError(text); msg != "" {
		t.Fatalf("%s:\n%s", msg, text)
	}
}

func TestScopedPostMergeCleanupOrderRejectsUnscopedAndPrematureClean(t *testing.T) {
	valid := confirmMergeFirst + ". Then run " + scopedGitCleanFd + ". " + pullMergedBranch + "."
	if msg := scopedPostMergeCleanupOrderError(valid); msg != "" {
		t.Fatalf("valid sequence rejected: %s", msg)
	}
	premature := scopedGitCleanFd + ". " + confirmMergeFirst + ". " + pullMergedBranch + "."
	if scopedPostMergeCleanupOrderError(premature) == "" {
		t.Fatal("expected rejection of cleanup before merge confirmation")
	}
	unscoped := confirmMergeFirst + ". Then run " + gitCleanFdLiteral + ". " + pullMergedBranch + "."
	if scopedPostMergeCleanupOrderError(unscoped) == "" {
		t.Fatal("expected rejection of unscoped cleanup")
	}
}

package worktree

import (
	"testing"
	"time"
)

func TestApplyListUpdatedDisplayUsesUserTimezoneAtMinutePrecision(t *testing.T) {
	t.Parallel()
	commitZone := time.FixedZone("commit", -7*60*60)
	userZone := time.FixedZone("user", 2*60*60)
	entries := []worktreeEntry{{
		lastUpdated: time.Date(2026, time.July, 26, 23, 58, 45, 123, commitZone),
		updatedText: "Jul 26, 2026",
	}}

	applyListUpdatedDisplay(entries, userZone)

	if got, want := entries[0].updatedText, "Jul 27, 2026 08:58"; got != want {
		t.Fatalf("updated text = %q, want %q", got, want)
	}
}

package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// workCleanupSyncStates lists every state the published primary checkout sync recipe can print.
var workCleanupSyncStates = []string{"UPDATED", "CURRENT", "KEEP_DIVERGED", "KEEP_UNCOMMITTED", "KEEP_DETACHED", "KEEP_REFUSED", "UNAVAILABLE"}

// workCleanupMergedLane returns a lane whose commit has merged into the remote default branch.
func workCleanupMergedLane(t *testing.T) (primary, lane string) {
	t.Helper()
	primary, lane = workCleanupLane(t)
	workCleanupGit(t, primary, "remote", "set-head", "origin", "main")
	workCleanupGit(t, lane, "push", "-q", "origin", "HEAD:main")
	return primary, lane
}

// workCleanupSync runs the published sync recipe from dir with extra environment entries.
func workCleanupSync(t *testing.T, dir string, env ...string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX recipe needs a shell")
	}
	recipe := contextEvidenceFence(t, workCleanupRule(t).Content, "sh", 1)
	cmd := exec.Command("sh", "-c", recipe)
	cmd.Dir = dir
	cmd.Env = append(append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1"), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync recipe failed: %v\n%s", err, output)
	}
	return strings.TrimSpace(string(output))
}

// TestWorkCleanupPrimarySyncStates exercises each sync state in disposable repositories.
func TestWorkCleanupPrimarySyncStates(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, primary, lane string) []string
		want    string
		moved   bool
	}{
		{name: "clean primary on default branch", want: "UPDATED", moved: true},
		{name: "unrelated untracked file", want: "UPDATED", moved: true, prepare: func(t *testing.T, primary, _ string) []string {
			writeFile(t, filepath.Join(primary, "scratch.txt"), "notes\n")
			return nil
		}},
		{name: "primary on another branch", want: "UPDATED", moved: true, prepare: func(t *testing.T, primary, _ string) []string {
			workCleanupGit(t, primary, "switch", "-q", "-c", "human-topic")
			return nil
		}},
		{name: "explicit base without remote HEAD", want: "UPDATED", moved: true, prepare: func(t *testing.T, primary, _ string) []string {
			workCleanupGit(t, primary, "remote", "set-head", "origin", "--delete")
			return []string{"base=main"}
		}},
		{name: "already current", want: "CURRENT", prepare: func(t *testing.T, primary, _ string) []string {
			workCleanupGit(t, primary, "pull", "-q", "--ff-only")
			return nil
		}},
		{name: "local commit on default branch", want: "KEEP_DIVERGED", prepare: func(t *testing.T, primary, _ string) []string {
			writeFile(t, filepath.Join(primary, "local.go"), "package main\n")
			workCleanupGit(t, primary, "add", "local.go")
			workCleanupGit(t, primary, "commit", "-q", "-m", "local only")
			return nil
		}},
		{name: "modified tracked file", want: "KEEP_UNCOMMITTED", prepare: func(t *testing.T, primary, _ string) []string {
			writeFile(t, filepath.Join(primary, "main.go"), "package main\n\nfunc f() {}\n")
			return nil
		}},
		{name: "detached head", want: "KEEP_DETACHED", prepare: func(t *testing.T, primary, _ string) []string {
			workCleanupGit(t, primary, "switch", "-q", "--detach")
			return nil
		}},
		{name: "untracked file in the way", want: "KEEP_REFUSED", prepare: func(t *testing.T, primary, _ string) []string {
			writeFile(t, filepath.Join(primary, "feature.go"), "human draft\n")
			return nil
		}},
		{name: "default branch in another worktree", want: "KEEP_REFUSED", prepare: func(t *testing.T, primary, _ string) []string {
			workCleanupGit(t, primary, "switch", "-q", "-c", "human-topic")
			workCleanupGit(t, primary, "worktree", "add", "-q", filepath.Join(t.TempDir(), "main"), "main")
			return nil
		}},
		{name: "no default branch known", want: "UNAVAILABLE", prepare: func(t *testing.T, primary, _ string) []string {
			workCleanupGit(t, primary, "remote", "set-head", "origin", "--delete")
			return nil
		}},
		{name: "unreachable remote", want: "UNAVAILABLE", prepare: func(t *testing.T, primary, _ string) []string {
			workCleanupGit(t, primary, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
			return nil
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary, lane := workCleanupMergedLane(t)
			var env []string
			if tt.prepare != nil {
				env = tt.prepare(t, primary, lane)
			}
			before := workCleanupGit(t, primary, "rev-parse", "refs/heads/main")
			branch := workCleanupGit(t, primary, "rev-parse", "--abbrev-ref", "HEAD")
			status := workCleanupGit(t, primary, "status", "--porcelain")
			if got := workCleanupSync(t, lane, env...); got != tt.want {
				t.Fatalf("sync = %q, want %q", got, tt.want)
			}
			after := workCleanupGit(t, primary, "rev-parse", "refs/heads/main")
			if tt.moved != (after != before) {
				t.Fatalf("main moved = %v, want %v", after != before, tt.moved)
			}
			if tt.moved && after != workCleanupGit(t, lane, "rev-parse", "HEAD") {
				t.Fatal("main did not reach the merged head")
			}
			if got := workCleanupGit(t, primary, "rev-parse", "--abbrev-ref", "HEAD"); got != branch {
				t.Fatalf("primary branch changed from %q to %q", branch, got)
			}
			if !tt.moved {
				if got := workCleanupGit(t, primary, "status", "--porcelain"); got != status {
					t.Fatalf("primary working tree changed:\n%s\nwant:\n%s", got, status)
				}
			}
		})
	}
}

// TestWorkCleanupPrimarySyncRunsFromPrimary verifies the recipe also works inside the primary checkout.
func TestWorkCleanupPrimarySyncRunsFromPrimary(t *testing.T) {
	primary, _ := workCleanupMergedLane(t)
	if got := workCleanupSync(t, primary); got != "UPDATED" {
		t.Fatalf("sync = %q, want UPDATED", got)
	}
	if got := workCleanupSync(t, primary); got != "CURRENT" {
		t.Fatalf("second sync = %q, want CURRENT", got)
	}
}

// TestWorkCleanupPrimarySyncDocumentsStates checks the rule explains every sync state.
func TestWorkCleanupPrimarySyncDocumentsStates(t *testing.T) {
	content := workCleanupRule(t).Content
	for _, state := range workCleanupSyncStates {
		if !strings.Contains(content, "`"+state+"`") {
			t.Errorf("rule does not explain sync state %s", state)
		}
	}
	for _, invariant := range []string{
		"It only fast-forwards the local default branch",
		"It never resets, stashes, rebases, or makes a merge commit",
		"keep the checkout as it is",
	} {
		if !strings.Contains(content, invariant) {
			t.Errorf("lost sync invariant: %s", invariant)
		}
	}
}

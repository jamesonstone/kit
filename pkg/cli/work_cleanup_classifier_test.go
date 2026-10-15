package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// workCleanupStates lists every state the published classifier can print.
var workCleanupStates = []string{"KEEP_LOCAL_STATE", "UNAVAILABLE"}

// workCleanupGit runs git in dir and fails the test on error.
func workCleanupGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

// workCleanupLane creates a bare remote, a primary clone, and a pushed lane worktree.
func workCleanupLane(t *testing.T) (primary, lane string) {
	t.Helper()
	base := t.TempDir()
	remote := filepath.Join(base, "remote.git")
	primary = filepath.Join(base, "primary")
	lane = filepath.Join(base, "worktrees", "GH-1")
	workCleanupGit(t, base, "init", "--bare", "-b", "main", remote)
	workCleanupGit(t, base, "clone", "-q", remote, primary)
	workCleanupGit(t, primary, "config", "user.name", "Test User")
	workCleanupGit(t, primary, "config", "user.email", "test@example.com")
	writeFile(t, filepath.Join(primary, ".gitignore"), ".env\n.envrc\nnode_modules/\n")
	writeFile(t, filepath.Join(primary, "main.go"), "package main\n")
	workCleanupGit(t, primary, "add", ".gitignore", "main.go")
	workCleanupGit(t, primary, "commit", "-q", "-m", "base")
	workCleanupGit(t, primary, "push", "-q", "origin", "main")
	workCleanupGit(t, primary, "worktree", "add", "-q", "-b", "GH-1", lane, "origin/main")
	writeFile(t, filepath.Join(lane, "feature.go"), "package main\n")
	workCleanupGit(t, lane, "add", "feature.go")
	workCleanupGit(t, lane, "commit", "-q", "-m", "feature")
	workCleanupGit(t, lane, "push", "-q", "origin", "GH-1")
	return primary, lane
}

// workCleanupClassify runs the published classifier recipe against one worktree.
func workCleanupClassify(t *testing.T, worktreePath string, env ...string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX recipe needs a shell")
	}
	recipe := contextEvidenceFence(t, workCleanupRule(t).Content, "sh", 0)
	cmd := exec.Command("sh", "-c", recipe)
	cmd.Dir = t.TempDir()
	cmd.Env = append(append(os.Environ(), "worktree_path="+worktreePath, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1"), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("classifier failed: %v\n%s", err, output)
	}
	return strings.TrimSpace(string(output))
}

// TestWorkCleanupClassifierStates exercises each classifier state in disposable repositories.
func TestWorkCleanupClassifierStates(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, primary, lane string) string
		want    string
	}{
		{name: "clean pushed lane", want: "KEEP_LOCAL_STATE"},
		{name: "regenerable ignored content and env symlinks", want: "KEEP_LOCAL_STATE", prepare: func(t *testing.T, primary, lane string) string {
			writeFile(t, filepath.Join(lane, "node_modules", "dep", "index.js"), "module.exports = 1\n")
			writeFile(t, filepath.Join(primary, ".env"), "TOKEN=fixture\n")
			if err := os.Symlink(filepath.Join(primary, ".env"), filepath.Join(lane, ".env")); err != nil {
				t.Fatal(err)
			}
			return lane
		}},
		{name: "squash merged with deleted remote branch", want: "KEEP_LOCAL_STATE", prepare: func(t *testing.T, primary, lane string) string {
			remote := workCleanupGit(t, primary, "remote", "get-url", "origin")
			head := workCleanupGit(t, lane, "rev-parse", "HEAD")
			workCleanupGit(t, remote, "update-ref", "refs/pull/1/head", head)
			workCleanupGit(t, primary, "push", "-q", "origin", "--delete", "GH-1")
			workCleanupGit(t, primary, "fetch", "-q", "--prune", "origin")
			return lane
		}},
		{name: "modified tracked file", want: "KEEP_LOCAL_STATE", prepare: func(t *testing.T, _, lane string) string {
			writeFile(t, filepath.Join(lane, "feature.go"), "package main\n\nfunc f() {}\n")
			return lane
		}},
		{name: "untracked file", want: "KEEP_LOCAL_STATE", prepare: func(t *testing.T, _, lane string) string {
			writeFile(t, filepath.Join(lane, "notes.txt"), "draft\n")
			return lane
		}},
		{name: "ignored real env file", want: "KEEP_LOCAL_STATE", prepare: func(t *testing.T, _, lane string) string {
			writeFile(t, filepath.Join(lane, ".env"), "TOKEN=fixture\n")
			return lane
		}},
		{name: "unpushed commit", want: "KEEP_LOCAL_STATE", prepare: func(t *testing.T, _, lane string) string {
			writeFile(t, filepath.Join(lane, "local.go"), "package main\n")
			workCleanupGit(t, lane, "add", "local.go")
			workCleanupGit(t, lane, "commit", "-q", "-m", "local only")
			return lane
		}},
		{name: "missing worktree", want: "UNAVAILABLE", prepare: func(t *testing.T, _, lane string) string {
			return lane + "-missing"
		}},
		{name: "unreachable remote", want: "KEEP_LOCAL_STATE", prepare: func(t *testing.T, primary, lane string) string {
			workCleanupGit(t, primary, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
			return lane
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary, lane := workCleanupLane(t)
			target := lane
			if tt.prepare != nil {
				target = tt.prepare(t, primary, lane)
			}
			before := workCleanupSnapshot(t, filepath.Dir(primary))
			if got := workCleanupClassify(t, target); got != tt.want {
				t.Fatalf("classifier = %q, want %q", got, tt.want)
			}
			assertWorkCleanupUnchanged(t, filepath.Dir(primary), before)
		})
	}
}

package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func workCleanupGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_NO_REPLACE_OBJECTS=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
func workCleanupLane(t *testing.T) (primary, lane string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("documented POSIX actions")
	}
	base := t.TempDir()
	remote := filepath.Join(base, "remote.git")
	primary = filepath.Join(base, "primary")
	lane = filepath.Join(base, "lane")
	workCleanupGit(t, base, "init", "--bare", "-q", "-b", "main", remote)
	workCleanupGit(t, base, "clone", "-q", remote, primary)
	workCleanupGit(t, primary, "config", "user.name", "Test User")
	workCleanupGit(t, primary, "config", "user.email", "test@example.invalid")
	writeFile(t, filepath.Join(primary, ".gitignore"), "node_modules/\n")
	writeFile(t, filepath.Join(primary, "source"), "base source\n")
	workCleanupGit(t, primary, "add", ".")
	workCleanupGit(t, primary, "commit", "-qm", "base")
	workCleanupGit(t, primary, "push", "-q", "origin", "main")
	workCleanupGit(t, primary, "worktree", "add", "-q", "-b", "GH-1", lane, "origin/main")
	writeFile(t, filepath.Join(lane, "feature"), "recoverable feature\n")
	workCleanupGit(t, lane, "add", "feature")
	workCleanupGit(t, lane, "commit", "-qm", "feature")
	workCleanupGit(t, lane, "push", "-q", "origin", "GH-1")
	return primary, lane
}

// Run only the published action, after fixture-specific evidence is established.
// These tests verify Git behavior; the rule's agent evidence gates are checked separately.
func workCleanupAction(t *testing.T, fence int, env ...string) (string, error) {
	t.Helper()
	recipe := contextEvidenceFence(t, workCleanupRule(t).Content, "sh", fence)
	cmd := exec.Command("sh", "-c", recipe)
	cmd.Dir = t.TempDir()
	cmd.Env = append(append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1"), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestWorkCleanupVerifiedRemovalAndRemoteRestore(t *testing.T) {
	primary, lane := workCleanupLane(t)
	head := workCleanupGit(t, lane, "rev-parse", "HEAD")
	live := workCleanupGit(t, primary, "ls-remote", "origin", "refs/heads/GH-1")
	if !strings.Contains(live, head) {
		t.Fatal("fixture has no source recovery")
	}
	if state := workCleanupGit(t, lane, "status", "--porcelain=v1", "--untracked-files=all", "--ignored=matching"); state != "" {
		t.Fatalf("fixture local state: %s", state)
	}
	if out, err := workCleanupAction(t, 0, "primary="+primary, "lane="+lane); err != nil {
		t.Fatalf("normal removal: %v %s", err, out)
	}
	if _, err := os.Stat(lane); !os.IsNotExist(err) {
		t.Fatal("eligible lane was not removed")
	}
	workCleanupGit(t, primary, "update-ref", "-d", "refs/heads/GH-1", head)
	restored := filepath.Join(t.TempDir(), "restore")
	remote := workCleanupGit(t, primary, "remote", "get-url", "origin")
	workCleanupGit(t, primary, "clone", "-q", "--branch", "GH-1", remote, restored)
	if got := workCleanupGit(t, restored, "rev-parse", "HEAD"); got != head {
		t.Fatal("remote source did not restore")
	}
	if got := readFile(t, filepath.Join(restored, "feature")); got != "recoverable feature\n" {
		t.Fatalf("restored bytes: %q", got)
	}
}

func TestWorkCleanupInventoryExposesIgnoredData(t *testing.T) {
	_, lane := workCleanupLane(t)
	writeFile(t, filepath.Join(lane, "node_modules", "data.db"), "unique local database")
	if got := workCleanupGit(t, lane, "status", "--porcelain"); got != "" {
		t.Fatalf("fixture ordinary status: %s", got)
	}
	inventory := workCleanupGit(t, lane, "status", "--porcelain=v1", "--untracked-files=all", "--ignored=matching")
	if !strings.Contains(inventory, "!! node_modules/") {
		t.Fatalf("ignored data absent: %q", inventory)
	}
	if got := readFile(t, filepath.Join(lane, "node_modules", "data.db")); got != "unique local database" {
		t.Fatal("inventory changed data")
	}
}

func TestWorkCleanupLiveRefsRejectStaleRecovery(t *testing.T) {
	primary, lane := workCleanupLane(t)
	head := workCleanupGit(t, lane, "rev-parse", "HEAD")
	workCleanupGit(t, primary, "push", "-q", "origin", "--delete", "GH-1")
	workCleanupGit(t, primary, "update-ref", "refs/remotes/origin/stale", head)
	live := workCleanupGit(t, primary, "ls-remote", "origin", "refs/heads/*", "refs/pull/*/head")
	if strings.Contains(live, head) {
		t.Fatal("deleted head unexpectedly live")
	}
	if got := workCleanupGit(t, primary, "rev-parse", "refs/remotes/origin/stale"); got != head {
		t.Fatal("fixture lacks stale cached recovery")
	}
	remote := workCleanupGit(t, primary, "remote", "get-url", "origin")
	workCleanupGit(t, remote, "update-ref", "refs/pull/1/head", head)
	if got := workCleanupGit(t, primary, "ls-remote", "origin", "refs/pull/*/head"); !strings.Contains(got, head) {
		t.Fatal("squash source head not discoverable")
	}
}

func TestWorkCleanupHiddenFlagsAndReflogNeedSeparateEvidence(t *testing.T) {
	_, lane := workCleanupLane(t)
	workCleanupGit(t, lane, "update-index", "--assume-unchanged", "source")
	writeFile(t, filepath.Join(lane, "source"), "hidden unique bytes")
	if got := workCleanupGit(t, lane, "status", "--porcelain"); got != "" {
		t.Fatalf("fixture did not hide edit: %s", got)
	}
	if flags := workCleanupGit(t, lane, "ls-files", "-v", "-f"); !strings.Contains(flags, "h source") {
		t.Fatalf("hidden flag absent: %s", flags)
	}
	head := workCleanupGit(t, lane, "rev-parse", "HEAD")
	unique := workCleanupGit(t, lane, "commit-tree", head+"^{tree}", "-p", head, "-m", "local-only reflog commit")
	workCleanupGit(t, lane, "update-ref", "refs/heads/GH-1", unique, head)
	workCleanupGit(t, lane, "update-ref", "refs/heads/GH-1", head, unique)
	if log := workCleanupGit(t, lane, "reflog", "show", "--format=%H", "GH-1"); !strings.Contains(log, unique) {
		t.Fatal("unique reflog fixture missing")
	}
	if live := workCleanupGit(t, lane, "ls-remote", "origin", "refs/heads/*", "refs/pull/*/head"); strings.Contains(live, unique) {
		t.Fatal("reflog commit unexpectedly recoverable")
	}
}

func TestWorkCleanupBranchDeletionRefusesChangedHead(t *testing.T) {
	primary, lane := workCleanupLane(t)
	head := workCleanupGit(t, lane, "rev-parse", "HEAD")
	workCleanupGit(t, primary, "worktree", "remove", lane)
	next := workCleanupGit(t, primary, "commit-tree", head+"^{tree}", "-p", head, "-m", "concurrent branch work")
	workCleanupGit(t, primary, "update-ref", "refs/heads/GH-1", next, head)
	cmd := exec.Command("git", "-C", primary, "update-ref", "-d", "refs/heads/GH-1", head)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("changed branch removed: %s", out)
	}
	if got := workCleanupGit(t, primary, "rev-parse", "GH-1"); got != next {
		t.Fatal("concurrent commit lost")
	}
}

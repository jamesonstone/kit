package cleanup

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type cleanupFixture struct {
	t                              *testing.T
	ctx                            context.Context
	base, repo, remote, lane, head string
	s                              *Service
	receipt                        Receipt
	options                        Options
}

func newCleanupFixture(t *testing.T) *cleanupFixture {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("POSIX cleanup fixture")
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", base)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "config"))
	f := &cleanupFixture{t: t, ctx: context.Background(), base: base, repo: filepath.Join(base, "primary"), remote: filepath.Join(base, "remote.git"), lane: filepath.Join(base, "lane"), s: New()}
	if err := os.Mkdir(f.repo, 0755); err != nil {
		t.Fatal(err)
	}
	f.git(f.repo, "init", "-q", "-b", "main")
	f.git(f.repo, "config", "user.name", "Fixture")
	f.git(f.repo, "config", "user.email", "fixture@example.invalid")
	mustTreeWrite(t, filepath.Join(f.repo, "file"), "recoverable content\n")
	mustTreeWrite(t, filepath.Join(f.repo, ".gitignore"), "ignored\n")
	f.git(f.repo, "add", ".")
	f.git(f.repo, "commit", "-q", "-m", "initial")
	f.git(f.repo, "init", "-q", "--bare", f.remote)
	f.git(f.repo, "remote", "add", "origin", f.remote)
	f.git(f.repo, "push", "-q", "-u", "origin", "main")
	f.git(f.remote, "symbolic-ref", "HEAD", "refs/heads/main")
	f.git(f.repo, "worktree", "add", "-q", "-b", "lane", f.lane)
	f.git(f.lane, "push", "-q", "-u", "origin", "lane")
	f.head = f.git(f.lane, "rev-parse", "HEAD")
	f.s.run = func(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
		if name == "lsof" {
			return []byte("p99999\nn/unrelated\n"), nil
		}
		return runCommand(ctx, cwd, name, args...)
	}
	f.receipt = f.record("worktree", f.lane, "")
	f.options = Options{Owner: "task", FinishedHead: f.head, StopEvidence: "current human stop", WritersStopped: true, Apply: true}
	return f
}
func (f *cleanupFixture) git(repo string, args ...string) string {
	f.t.Helper()
	out, err := treeTestRun(f.ctx, repo, "git", args...)
	if err != nil {
		f.t.Fatalf("git %v: %s: %v", args, out, err)
	}
	return strings.TrimSpace(string(out))
}
func (f *cleanupFixture) record(kind, path, ref string) Receipt {
	f.t.Helper()
	r, err := f.s.Record(f.ctx, kind, path, f.repo, "task", ref)
	if err != nil {
		f.t.Fatal(err)
	}
	return r
}
func (f *cleanupFixture) execute() Result { return f.s.Execute(f.ctx, f.receipt, f.options) }
func assertCleanupState(t *testing.T, r Result, want string) {
	t.Helper()
	if r.State != want {
		t.Fatalf("state %s, want %s: %s", r.State, want, r.Reason)
	}
}
func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Fatal(err)
	}
}
func assertAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("target still exists or inaccessible: %s: %v", path, err)
	}
}

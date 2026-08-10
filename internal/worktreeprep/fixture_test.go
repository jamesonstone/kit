package worktreeprep

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type repositoryFixture struct {
	root       string
	home       string
	primary    string
	remote     string
	repository string
	preparer   *Preparer
}

func newRepositoryFixture(t *testing.T) repositoryFixture {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "owner", "project.git")
	primary := filepath.Join(root, "primary")
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Dir(remote), 0o755); err != nil {
		t.Fatal(err)
	}
	gitCommand(t, root, "init", "--bare", remote)
	gitCommand(t, root, "init", "-b", "main", primary)
	gitCommand(t, primary, "config", "user.name", "Test User")
	gitCommand(t, primary, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(primary, "README.md"), []byte("fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommand(t, primary, "add", "README.md")
	gitCommand(t, primary, "commit", "-m", "initial")
	gitCommand(t, primary, "remote", "add", "origin", remote)
	gitCommand(t, primary, "push", "-u", "origin", "main")
	gitCommand(t, remote, "symbolic-ref", "HEAD", "refs/heads/main")
	preparer := New()
	preparer.homeDir = func() (string, error) { return home, nil }
	return repositoryFixture{
		root:       root,
		home:       home,
		primary:    primary,
		remote:     remote,
		repository: "owner/project",
		preparer:   preparer,
	}
}

func (fixture repositoryFixture) createLocalBranch(t *testing.T, branch string) {
	t.Helper()
	gitCommand(t, fixture.primary, "branch", branch)
}

func (fixture repositoryFixture) createRemoteBranch(t *testing.T, branch string) {
	t.Helper()
	fixture.createLocalBranch(t, branch)
	gitCommand(t, fixture.primary, "push", "origin", branch)
	gitCommand(t, fixture.primary, "branch", "-d", branch)
}

func gitCommand(t *testing.T, cwd string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = cwd
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func assertSymlink(t *testing.T, path, target string) {
	t.Helper()
	actual, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("readlink %s: %v", path, err)
	}
	if !samePath(actual, target) {
		t.Fatalf("symlink %s target = %q, want %q", path, actual, target)
	}
}

package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/worktree"
)

func TestPreparePRRepairContextRecordsDirtyWorktreeDecision(t *testing.T) {
	for _, test := range []struct {
		name        string
		answer      string
		disposition repairChangeDisposition
	}{
		{name: "include", answer: "yes\n", disposition: repairChangesInclude},
		{name: "exclude", answer: "\n", disposition: repairChangesExclude},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := newRepairContextRepository(t, "jamesonstone/kit", "GH-67")
			if err := os.WriteFile(filepath.Join(repo, "dirty.txt"), []byte("preserve\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			headOID := repairContextGit(t, repo, "rev-parse", "HEAD")

			previousPrepare := preparePullRequestWorktree
			preparePullRequestWorktree = func(
				_ context.Context,
				cwd string,
				number int,
			) (worktree.PullRequestRepair, error) {
				if cwd != repo || number != 67 {
					t.Fatalf("prepare target = %s#%d", cwd, number)
				}
				return worktree.PullRequestRepair{
					PreparedWorktree: worktree.PreparedWorktree{
						Path:   repo,
						Branch: "GH-67",
					},
					Repository: "jamesonstone/kit",
					Number:     67,
					URL:        "https://github.com/jamesonstone/kit/pull/67",
					HeadRefOID: headOID,
				}, nil
			}
			t.Cleanup(func() { preparePullRequestWorktree = previousPrepare })

			var prompt bytes.Buffer
			repair, err := preparePRRepairContext(
				context.Background(),
				strings.NewReader(test.answer),
				&prompt,
				repo,
				"jamesonstone/kit#67",
			)
			if err != nil {
				t.Fatalf("preparePRRepairContext() error = %v", err)
			}
			if repair.ExistingChanges != test.disposition {
				t.Fatalf("ExistingChanges = %q, want %q", repair.ExistingChanges, test.disposition)
			}
			if repair.WorktreePath != repo ||
				repair.HeadBranch != "GH-67" ||
				repair.LocalHeadOID != headOID ||
				repair.ExpectedHeadOID != headOID ||
				repair.PushTarget != "origin/GH-67" {
				t.Fatalf("unexpected repair context: %#v", repair)
			}
			if !strings.Contains(repair.DirtyStatus, "?? dirty.txt") {
				t.Fatalf("dirty status missing untracked file: %q", repair.DirtyStatus)
			}
			if !strings.Contains(prompt.String(), "Include these changes in the existing pull request repair? [y/N]") {
				t.Fatalf("dirty confirmation missing:\n%s", prompt.String())
			}
		})
	}
}

func TestPreparePRRepairContextRejectsAnotherRepository(t *testing.T) {
	repo := newRepairContextRepository(t, "other/project", "GH-67")
	previousPrepare := preparePullRequestWorktree
	preparePullRequestWorktree = func(
		context.Context,
		string,
		int,
	) (worktree.PullRequestRepair, error) {
		t.Fatal("repository mismatch must fail before worktree preparation")
		return worktree.PullRequestRepair{}, nil
	}
	t.Cleanup(func() { preparePullRequestWorktree = previousPrepare })

	_, err := preparePRRepairContext(
		context.Background(),
		strings.NewReader(""),
		io.Discard,
		repo,
		"jamesonstone/kit#67",
	)
	if err == nil || !strings.Contains(err.Error(), "does not belong to current clone other/project") {
		t.Fatalf("expected repository mismatch, got %v", err)
	}
}

func TestConfirmRepairChangesRequiresAnExplicitResponseAtEOF(t *testing.T) {
	_, err := confirmRepairChanges(strings.NewReader(""), io.Discard, repairContext{
		WorktreePath:      "/tmp/kit/GH-67",
		TargetDescription: "https://github.com/jamesonstone/kit/pull/67",
		PRURL:             "https://github.com/jamesonstone/kit/pull/67",
		DirtyStatus:       " M pkg/cli/pr.go",
	})
	if err == nil || !strings.Contains(err.Error(), "requires an explicit y or n response") {
		t.Fatalf("expected explicit-response error, got %v", err)
	}
}

func TestPrepareCIRepairContextPrefersOpenPullRequest(t *testing.T) {
	previousOutput := repairContextCommandOutput
	previousResolver := resolvePRRepairContext
	t.Cleanup(func() {
		repairContextCommandOutput = previousOutput
		resolvePRRepairContext = previousResolver
	})

	repairContextCommandOutput = func(_ string, name string, args ...string) ([]byte, error) {
		if name != "gh" || len(args) < 2 || args[0] != "pr" || args[1] != "list" {
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		return []byte(`[{"number":67,"url":"https://github.com/jamesonstone/kit/pull/67"}]`), nil
	}
	want := &repairContext{PRNumber: 67, WorktreePath: "/tmp/kit/GH-67"}
	resolvePRRepairContext = func(
		_ context.Context,
		_ io.Reader,
		_ io.Writer,
		cwd string,
		prRef string,
	) (*repairContext, error) {
		if cwd != "/tmp/kit" || prRef != "https://github.com/jamesonstone/kit/pull/67" {
			t.Fatalf("resolved PR repair target = cwd %q, ref %q", cwd, prRef)
		}
		return want, nil
	}

	got, err := prepareCIRepairContext(
		context.Background(),
		strings.NewReader(""),
		io.Discard,
		"/tmp/kit",
		ciDiagnosis{Target: ciTarget{
			Repository: "jamesonstone/kit",
			Branch:     "GH-67",
		}},
	)
	if err != nil {
		t.Fatalf("prepareCIRepairContext() error = %v", err)
	}
	if got != want {
		t.Fatalf("repair context = %#v, want %#v", got, want)
	}
}

func TestPrepareCIRepairContextRefusesDefaultBranchWithoutPullRequest(t *testing.T) {
	previousOutput := repairContextCommandOutput
	t.Cleanup(func() { repairContextCommandOutput = previousOutput })
	repairContextCommandOutput = func(_ string, name string, args ...string) ([]byte, error) {
		switch {
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return []byte(`[]`), nil
		case name == "gh" && len(args) >= 2 && args[0] == "repo" && args[1] == "view":
			return []byte("main\n"), nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return nil, nil
		}
	}

	_, err := prepareCIRepairContext(
		context.Background(),
		strings.NewReader(""),
		io.Discard,
		"/tmp/kit",
		ciDiagnosis{Target: ciTarget{
			Repository: "jamesonstone/kit",
			Branch:     "main",
		}},
	)
	if err == nil || !strings.Contains(err.Error(), "protected default branch") {
		t.Fatalf("expected protected default-branch refusal, got %v", err)
	}
}

func TestResolvePromptWorktreeRootNormalizesNestedDirectory(t *testing.T) {
	repo := newRepairContextRepository(t, "jamesonstone/kit", "GH-67")
	nested := filepath.Join(repo, "nested", "directory")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := resolvePromptWorktreeRoot(nested)
	if err != nil {
		t.Fatalf("resolvePromptWorktreeRoot() error = %v", err)
	}
	rootInfo, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	repoInfo, err := os.Stat(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(rootInfo, repoInfo) {
		t.Fatalf("root = %q, want same directory as %q", root, repo)
	}
}

func newRepairContextRepository(t *testing.T, remote, branch string) string {
	t.Helper()
	repo := t.TempDir()
	repairContextGit(t, repo, "init", "-b", branch)
	repairContextGit(t, repo, "config", "user.name", "Kit Test")
	repairContextGit(t, repo, "config", "user.email", "kit@example.com")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("tracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repairContextGit(t, repo, "add", "tracked.txt")
	repairContextGit(t, repo, "commit", "-m", "initial")
	repairContextGit(t, repo, "remote", "add", "origin", "git@github.com:"+remote+".git")
	return repo
}

func repairContextGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

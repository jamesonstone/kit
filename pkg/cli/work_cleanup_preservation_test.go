package cli

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	stdreflect "reflect"
	"testing"
)

// workCleanupSnapshot includes primary/lane bytes, modes, symlinks and all Git metadata.
func workCleanupSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	state := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		value := fmt.Sprintf("%v", info.Mode())
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			value += ":" + target
		} else if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += fmt.Sprintf(":%x", sha256.Sum256(data))
		}
		state[rel] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func assertWorkCleanupUnchanged(t *testing.T, root string, before map[string]string) {
	t.Helper()
	if after := workCleanupSnapshot(t, root); !stdreflect.DeepEqual(before, after) {
		for path, value := range before {
			if after[path] != value {
				t.Errorf("local state changed: %s", path)
			}
		}
		for path := range after {
			if _, exists := before[path]; !exists {
				t.Errorf("inventory wrote %s", path)
			}
		}
	}
}

// TestWorkCleanupPreservesLocalState exercises actual local state, never user checkouts.
func TestWorkCleanupPreservesLocalState(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, primary, lane string)
	}{
		{"ignored database and credentials", func(t *testing.T, primary, lane string) {
			writeFile(t, filepath.Join(primary, ".git", "info", "exclude"), "*.db\n*.pem\n")
			writeFile(t, filepath.Join(lane, "local.db"), "unique database\n")
			writeFile(t, filepath.Join(lane, "credential.pem"), "synthetic credential\n")
		}},
		{"assume unchanged edit", func(t *testing.T, _, lane string) {
			workCleanupGit(t, lane, "update-index", "--assume-unchanged", "feature.go")
			writeFile(t, filepath.Join(lane, "feature.go"), "unique hidden bytes\n")
		}},
		{"skip worktree edit", func(t *testing.T, _, lane string) {
			workCleanupGit(t, lane, "update-index", "--skip-worktree", "feature.go")
			writeFile(t, filepath.Join(lane, "feature.go"), "unique skipped bytes\n")
		}},
		{"local reflog only commit", func(t *testing.T, _, lane string) {
			head := workCleanupGit(t, lane, "rev-parse", "HEAD")
			writeFile(t, filepath.Join(lane, "local.txt"), "reflog-only bytes\n")
			workCleanupGit(t, lane, "add", "local.txt")
			workCleanupGit(t, lane, "commit", "-q", "-m", "local")
			// Reset only this disposable fixture to reproduce the reviewed loss case.
			workCleanupGit(t, lane, "reset", "--hard", head)
		}},
		{"worktree config", func(t *testing.T, primary, lane string) {
			workCleanupGit(t, primary, "config", "extensions.worktreeConfig", "true")
			workCleanupGit(t, lane, "config", "--worktree", "fixture.unique", "local")
		}},
		{"paused git operation", func(t *testing.T, _, lane string) {
			gitdir := workCleanupGit(t, lane, "rev-parse", "--absolute-git-dir")
			writeFile(t, filepath.Join(gitdir, "rebase-merge", "head-name"), "refs/heads/GH-1\n")
			writeFile(t, filepath.Join(gitdir, "MERGE_HEAD"), workCleanupGit(t, lane, "rev-parse", "HEAD")+"\n")
		}},
		{"ignored executable mode", func(t *testing.T, _, lane string) {
			path := filepath.Join(lane, "node_modules", "local-tool")
			writeFile(t, path, "#!/bin/sh\nexit 0\n")
			if err := os.Chmod(path, 0o751); err != nil {
				t.Fatal(err)
			}
		}},
		{"active process", func(t *testing.T, _, lane string) {
			process := exec.Command("sleep", "30")
			process.Dir = lane
			if err := process.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = process.Process.Kill(); _ = process.Wait() })
		}},
		{"primary ignored collision with remote", func(t *testing.T, primary, lane string) {
			writeFile(t, filepath.Join(primary, ".git", "info", "exclude"), "collision\n")
			writeFile(t, filepath.Join(primary, "collision"), "unique primary bytes\n")
			writeFile(t, filepath.Join(lane, "collision"), "upstream bytes\n")
			workCleanupGit(t, lane, "add", "-f", "collision")
			workCleanupGit(t, lane, "commit", "-q", "-m", "remote collision")
			workCleanupGit(t, lane, "push", "-q", "origin", "HEAD:main")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary, lane := workCleanupLane(t)
			tt.prepare(t, primary, lane)
			root := filepath.Dir(primary)
			before := workCleanupSnapshot(t, root)
			for _, target := range []string{lane, primary} {
				if got := workCleanupClassify(t, target); got != "KEEP_LOCAL_STATE" {
					t.Fatalf("inventory = %s", got)
				}
			}
			assertWorkCleanupUnchanged(t, root, before)
		})
	}
}

// TestWorkCleanupUnknownChecksCannotAuthorizeRemoval has no dependency on lsof.
func TestWorkCleanupUnknownChecksCannotAuthorizeRemoval(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"missing lsof", "failed lsof", "failed git"} {
		t.Run(mode, func(t *testing.T) {
			primary, lane := workCleanupLane(t)
			tools := t.TempDir()
			if mode == "failed git" {
				writeFile(t, filepath.Join(tools, "git"), "#!/bin/sh\nexit 1\n")
				if err := os.Chmod(filepath.Join(tools, "git"), 0o755); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Symlink(git, filepath.Join(tools, "git")); err != nil {
				t.Fatal(err)
			}
			if mode == "failed lsof" {
				writeFile(t, filepath.Join(tools, "lsof"), "#!/bin/sh\nexit 2\n")
				if err := os.Chmod(filepath.Join(tools, "lsof"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			before := workCleanupSnapshot(t, filepath.Dir(primary))
			want := "KEEP_LOCAL_STATE"
			if mode == "failed git" {
				want = "UNAVAILABLE"
			}
			// Resolve sh before setting PATH so tools can be genuinely unavailable.
			if got := workCleanupClassify(t, lane, "PATH="+tools); got != want {
				t.Fatalf("inventory = %s, want %s", got, want)
			}
			assertWorkCleanupUnchanged(t, filepath.Dir(primary), before)
		})
	}
}

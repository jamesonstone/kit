package cleanup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueOrUnknownStateIsPreserved(t *testing.T) {
	cases := map[string]func(*cleanupFixture){
		"clean-assume-unchanged": func(f *cleanupFixture) { f.git(f.lane, "update-index", "--assume-unchanged", "file") },
		"clean-skip-worktree":    func(f *cleanupFixture) { f.git(f.lane, "update-index", "--skip-worktree", "file") },
		"file-permissions": func(f *cleanupFixture) {
			if err := os.Chmod(filepath.Join(f.lane, "file"), 0600); err != nil {
				t.Fatal(err)
			}
		},
		"root-permissions": func(f *cleanupFixture) {
			if err := os.Chmod(f.lane, 0700); err != nil {
				t.Fatal(err)
			}
		},
		"control-permissions": func(f *cleanupFixture) {
			if err := os.Chmod(filepath.Join(f.lane, ".git"), 0600); err != nil {
				t.Fatal(err)
			}
		},

		"ignored":   func(f *cleanupFixture) { mustTreeWrite(t, filepath.Join(f.lane, "ignored"), "unique") },
		"untracked": func(f *cleanupFixture) { mustTreeWrite(t, filepath.Join(f.lane, "untracked"), "unique") },
		"hidden-assume-unchanged": func(f *cleanupFixture) {
			f.git(f.lane, "update-index", "--assume-unchanged", "file")
			mustTreeWrite(t, filepath.Join(f.lane, "file"), "hidden")
		},
		"hidden-skip-worktree": func(f *cleanupFixture) {
			f.git(f.lane, "update-index", "--skip-worktree", "file")
			mustTreeWrite(t, filepath.Join(f.lane, "file"), "hidden")
		},
		"reflog-only-commit": func(f *cleanupFixture) {
			mustTreeWrite(t, filepath.Join(f.lane, "file"), "unique commit")
			f.git(f.lane, "add", "file")
			f.git(f.lane, "commit", "-q", "-m", "unique")
			f.git(f.lane, "reset", "-q", "--hard", f.head)
		},
		"worktree-config": func(f *cleanupFixture) {
			mustTreeWrite(t, filepath.Join(f.receipt.GitDir, "config.worktree"), "[custom]\n value = unique\n")
		},
		"branch-config":    func(f *cleanupFixture) { f.git(f.repo, "config", "branch.lane.description", "unique configuration") },
		"active-operation": func(f *cleanupFixture) { mustTreeWrite(t, filepath.Join(f.receipt.GitDir, "MERGE_HEAD"), f.head) },
		"wrong-owner":      func(f *cleanupFixture) { f.options.Owner = "other" },
		"no-handoff":       func(f *cleanupFixture) { f.options.WritersStopped = false },
		"no-completion":    func(f *cleanupFixture) { f.options.StopEvidence = "" },
		"identity-replacement": func(f *cleanupFixture) {
			if err := os.Rename(f.lane, f.lane+"-retained"); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(f.lane, 0755); err != nil {
				t.Fatal(err)
			}
			mustTreeWrite(t, filepath.Join(f.lane, "replacement"), "unique")
		},
		"remote-ref-deleted": func(f *cleanupFixture) {
			// Keep the cached tracking ref to demonstrate it cannot establish live recovery.
			f.git(f.repo, "update-ref", "refs/remotes/origin/lane", f.head)
			f.git(f.remote, "update-ref", "-d", "refs/heads/lane")
			f.git(f.remote, "update-ref", "-d", "refs/heads/main")
		},
		"concurrent-write": func(f *cleanupFixture) {
			f.s.beforeRevalidate = func() { mustTreeWrite(t, filepath.Join(f.lane, "late"), "new unique data") }
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f := newCleanupFixture(t)
			mutate(f)
			contents := preservedFiles(t, f.lane)
			r := f.execute()
			if r.State == "REMOVED" || r.State == "PRUNED" || r.State == "UPDATED" {
				t.Fatalf("unsafe cleanup: %+v", r)
			}
			assertExists(t, f.lane)
			for path, expected := range contents {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != expected {
					t.Fatalf("unique content changed: %s: %v", path, err)
				}
			}
			if name == "concurrent-write" {
				data, err := os.ReadFile(filepath.Join(f.lane, "late"))
				if err != nil || string(data) != "new unique data" {
					t.Fatal("concurrent data lost")
				}
			}
		})
	}
}
func TestUnavailableOrActiveProcessCheckPreservesTarget(t *testing.T) {
	for _, mode := range []string{"missing", "failed", "empty", "malformed", "active"} {
		t.Run(mode, func(t *testing.T) {
			f := newCleanupFixture(t)
			f.s.run = func(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
				if name != "lsof" {
					return runCommand(ctx, cwd, name, args...)
				}
				switch mode {
				case "missing":
					return nil, os.ErrNotExist
				case "failed":
					return nil, errors.New("permission denied")
				case "empty":
					return nil, nil
				case "malformed":
					return []byte("not an inventory"), nil
				default:
					return []byte("p99999\nn" + filepath.Join(f.lane, "file") + "\n"), nil
				}
			}
			r := f.execute()
			want := "UNAVAILABLE"
			if mode == "active" {
				want = "KEEP_IN_USE"
			}
			assertCleanupState(t, r, want)
			assertExists(t, f.lane)
		})
	}
}
func TestPRCompletionMustMatchExactFinishedHead(t *testing.T) {
	for _, tc := range []struct {
		name, state, head string
		want              string
	}{
		{"merged", "MERGED", "current", "REMOVED"}, {"closed", "CLOSED", "current", "REMOVED"}, {"open", "OPEN", "current", "KEEP_COMPLETION"}, {"stale-head", "MERGED", "stale", "KEEP_COMPLETION"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCleanupFixture(t)
			head := tc.head
			if head == "current" {
				head = f.head
			}
			f.options.PR = 42
			original := f.s.run
			f.s.run = func(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
				if name == "gh" {
					return []byte(fmt.Sprintf(`{"state":%q,"headRefOid":%q}`, tc.state, head)), nil
				}
				return original(ctx, cwd, name, args...)
			}
			assertCleanupState(t, f.execute(), tc.want)
			if tc.want != "REMOVED" {
				assertExists(t, f.lane)
			}
		})
	}
}
func TestGeneratedProvenanceAndUnsupportedKindsPreserve(t *testing.T) {
	for _, mode := range []string{"changed", "nested-git", "unsupported"} {
		t.Run(mode, func(t *testing.T) {
			f := newCleanupFixture(t)
			path := filepath.Join(f.base, "generated")
			if err := os.Mkdir(path, 0755); err != nil {
				t.Fatal(err)
			}
			mustTreeWrite(t, filepath.Join(path, "payload"), "generated")
			f.receipt = f.record("cache", path, "")
			switch mode {
			case "changed":
				mustTreeWrite(t, filepath.Join(path, "payload"), "unique change")
			case "nested-git":
				if err := os.Mkdir(filepath.Join(path, ".git"), 0755); err != nil {
					t.Fatal(err)
				}
			case "unsupported":
				f.receipt.Kind = "docker-system"
			}
			r := f.execute()
			if r.State == "REMOVED" {
				t.Fatalf("unsafe generated cleanup: %+v", r)
			}
			assertExists(t, path)
		})
	}
}

// Read only ordinary fixture files; never follow a symlink into unrelated state.
func preservedFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[path] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

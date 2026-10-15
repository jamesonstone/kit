package cleanup

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func treeTestRun(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = cwd
	return cmd.Output()
}

func treeFixture(t *testing.T, format string) (string, string) {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("POSIX filesystem proof")
	}
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	lane := filepath.Join(base, "lane")
	if err := os.Mkdir(repo, 0755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		if out, err := treeTestRun(context.Background(), repo, "git", args...); err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
	}
	git("init", "--object-format="+format)
	git("config", "user.name", "Fixture")
	git("config", "user.email", "fixture@example.invalid")
	for name, data := range map[string]string{"file": "original\n", ".gitignore": "ignored\n"} {
		if err := os.WriteFile(filepath.Join(repo, name), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("file", filepath.Join(repo, "link")); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-m", "fixture")
	git("worktree", "add", "-b", "lane", lane)
	return repo, lane
}

func TestTrackedTreeMatchesPhysicalObjects(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			_, lane := treeFixture(t, format)
			first, err := verifyTrackedTree(context.Background(), treeTestRun, lane, "HEAD")
			if err != nil {
				t.Fatal(err)
			}
			second, err := snapshotTree(lane)
			if err != nil {
				t.Fatal(err)
			}
			if first != second || first.Bytes == 0 || first.Digest == "" {
				t.Fatalf("unstable or empty snapshot: %+v %+v", first, second)
			}
		})
	}
}

func TestTrackedTreeRejectsUniquePhysicalState(t *testing.T) {
	cases := map[string]func(*testing.T, string){
		"ignored":   func(t *testing.T, p string) { mustTreeWrite(t, filepath.Join(p, "ignored"), "unique") },
		"untracked": func(t *testing.T, p string) { mustTreeWrite(t, filepath.Join(p, "extra"), "unique") },
		"empty-directory": func(t *testing.T, p string) {
			if err := os.Mkdir(filepath.Join(p, "empty"), 0755); err != nil {
				t.Fatal(err)
			}
		},
		"hidden-assume-unchanged": func(t *testing.T, p string) {
			treeGit(t, p, "update-index", "--assume-unchanged", "file")
			mustTreeWrite(t, filepath.Join(p, "file"), "hidden")
		},
		"hidden-skip-worktree": func(t *testing.T, p string) {
			treeGit(t, p, "update-index", "--skip-worktree", "file")
			mustTreeWrite(t, filepath.Join(p, "file"), "hidden")
		},
		"non-Git-permission": func(t *testing.T, p string) {
			if err := os.Chmod(filepath.Join(p, "file"), 0600); err != nil {
				t.Fatal(err)
			}
		},
		"hardlink": func(t *testing.T, p string) {
			if err := os.Link(filepath.Join(p, "file"), filepath.Join(t.TempDir(), "copy")); err != nil {
				t.Fatal(err)
			}
		},
		"missing": func(t *testing.T, p string) {
			if err := os.Remove(filepath.Join(p, "file")); err != nil {
				t.Fatal(err)
			}
		},
		"changed-symlink": func(t *testing.T, p string) {
			if err := os.Remove(filepath.Join(p, "link")); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("other", filepath.Join(p, "link")); err != nil {
				t.Fatal(err)
			}
		},
		"nested-repo": func(t *testing.T, p string) { treeGit(t, p, "init", "nested") },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			_, lane := treeFixture(t, "sha1")
			if _, err := verifyTrackedTree(context.Background(), treeTestRun, lane, "HEAD"); err != nil {
				t.Fatalf("invalid baseline fixture: %v", err)
			}
			mutate(t, lane)
			if _, err := verifyTrackedTree(context.Background(), treeTestRun, lane, "HEAD"); err == nil {
				t.Fatal("accepted unique local state")
			}
		})
	}
}

func TestSnapshotDoesNotFollowSymlinks(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("POSIX filesystem proof")
	}
	root := t.TempDir()
	outside := t.TempDir()
	mustTreeWrite(t, filepath.Join(outside, "secret"), "before")
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	first, err := snapshotTree(root)
	if err != nil {
		t.Fatal(err)
	}
	mustTreeWrite(t, filepath.Join(outside, "secret"), "after")
	second, err := snapshotTree(root)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("followed external symlink")
	}
	if _, err := snapshotTree(filepath.Join(root, "link")); err == nil {
		t.Fatal("accepted symlink root")
	}
}

func TestSnapshotRejectsDarwinExtendedMetadata(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin metadata fixture")
	}
	for _, kind := range []string{"xattr", "ACL", "flag"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "file")
			mustTreeWrite(t, path, "data")
			var cmd *exec.Cmd
			switch kind {
			case "xattr":
				cmd = exec.Command("/usr/bin/xattr", "-w", "user.kit-test", "value", path)
			case "ACL":
				cmd = exec.Command("/bin/chmod", "+a", "everyone allow read", path)
			case "flag":
				cmd = exec.Command("/usr/bin/chflags", "hidden", path)
			}
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("metadata fixture: %s: %v", out, err)
			}
			if _, err := snapshotTree(root); err == nil {
				t.Fatal("accepted extended metadata")
			}
		})
	}
}

func treeGit(t *testing.T, root string, args ...string) {
	t.Helper()
	if out, err := treeTestRun(context.Background(), root, "git", args...); err != nil {
		t.Fatalf("git %v: %s %v", args, out, err)
	}
}
func mustTreeWrite(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestTrackedTreePrimaryControlDirectoryExcluded(t *testing.T) {
	repo, _ := treeFixture(t, "sha1")
	path := filepath.Join(repo, ".git", "test-opaque-state")
	mustTreeWrite(t, path, "unique Git state checked separately by caller")
	// Directory timestamps still bind its identity; take the baseline after adding
	// the entry, then mutate only its bytes to prove contents are excluded.
	first, err := verifyTrackedTree(context.Background(), treeTestRun, repo, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	mustTreeWrite(t, path, "changed Git state")
	second, err := verifyTrackedTree(context.Background(), treeTestRun, repo, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("primary tree proof included Git control contents")
	}
}

func TestDarwinOSProvenanceIsOnlyXattrException(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin provenance policy")
	}
	root := t.TempDir()
	mustTreeWrite(t, filepath.Join(root, "file"), "data")
	// This host applies com.apple.provenance automatically, including to the root.
	// No attributes are stripped from this positive fixture.
	if _, err := snapshotTree(root); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotStreamsLargeGeneratedFile(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("POSIX streaming fixture")
	}
	root := t.TempDir()
	path := filepath.Join(root, "large-cache")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	const size int64 = 64 << 20
	if err := file.Truncate(size); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	first, err := snapshotTree(root)
	if err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&after)
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 8<<20 {
		t.Fatalf("snapshot allocated %d bytes for sparse file; expected bounded streaming", allocated)
	}
	if first.Bytes != size {
		t.Fatalf("logical bytes=%d, want %d", first.Bytes, size)
	}
	file, err = os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteAt([]byte{1}, size-1); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := snapshotTree(root)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digest == second.Digest {
		t.Fatal("streamed snapshot missed changed final byte")
	}
}

func TestOversizedGitControlFileFailsClosed(t *testing.T) {
	_, lane := treeFixture(t, "sha1")
	path := filepath.Join(lane, ".git")
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxGitControlBytes + 1); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshotTree(lane); err == nil {
		t.Fatal("accepted oversized Git control file")
	}
}

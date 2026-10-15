package cli

import (
	"path/filepath"
	"testing"
)

func TestWorkCleanupPrimaryFastForwardAndIgnoredCollision(t *testing.T) {
	for _, collision := range []bool{false, true} {
		t.Run(map[bool]string{false: "clean", true: "ignored collision"}[collision], func(t *testing.T) {
			primary, lane := workCleanupLane(t)
			path := filepath.Join("node_modules", "result")
			writeFile(t, filepath.Join(lane, path), "remote source")
			workCleanupGit(t, lane, "add", "-f", path)
			workCleanupGit(t, lane, "commit", "-qm", "track formerly ignored path")
			head := workCleanupGit(t, lane, "rev-parse", "HEAD")
			workCleanupGit(t, lane, "push", "-q", "origin", "HEAD:main")
			before := workCleanupGit(t, primary, "rev-parse", "HEAD")
			if collision {
				writeFile(t, filepath.Join(primary, path), "unique local result")
			}
			out, err := workCleanupAction(t, 1, "primary="+primary, "remote_head="+head)
			if collision {
				if err == nil {
					t.Fatalf("ignored collision overwritten: %s", out)
				}
				if got := readFile(t, filepath.Join(primary, path)); got != "unique local result" {
					t.Fatalf("local bytes changed: %q", got)
				}
				if got := workCleanupGit(t, primary, "rev-parse", "HEAD"); got != before {
					t.Fatal("refused sync changed default")
				}
			} else {
				if err != nil {
					t.Fatalf("eligible sync failed: %v %s", err, out)
				}
				if got := workCleanupGit(t, primary, "rev-parse", "HEAD"); got != head {
					t.Fatal("default did not advance")
				}
			}
		})
	}
}

func TestWorkCleanupPrimaryOtherBranchLeavesWorkingFilesUntouched(t *testing.T) {
	primary, lane := workCleanupLane(t)
	old := workCleanupGit(t, primary, "rev-parse", "main")
	head := workCleanupGit(t, lane, "rev-parse", "HEAD")
	workCleanupGit(t, lane, "push", "-q", "origin", "HEAD:main")
	workCleanupGit(t, primary, "switch", "-q", "-c", "human-work")
	writeFile(t, filepath.Join(primary, "source"), "human working files")
	workCleanupGit(t, primary, "update-ref", "refs/heads/main", head, old)
	if got := readFile(t, filepath.Join(primary, "source")); got != "human working files" {
		t.Fatal("default ref update changed active files")
	}
	if got := workCleanupGit(t, primary, "branch", "--show-current"); got != "human-work" {
		t.Fatal("active branch changed")
	}
	if got := workCleanupGit(t, primary, "rev-parse", "main"); got != head {
		t.Fatal("default not advanced")
	}
}

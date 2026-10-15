package cleanup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckoutFilterProofDistinguishesConfiguredAndActive(t *testing.T) {
	f := newCleanupFixture(t)
	f.git(f.repo, "config", "filter.fixture.smudge", "touch filter-executed")
	if err := f.s.checkoutFilters(f.ctx, f.repo, f.head); err != nil {
		t.Fatalf("unused filter is not executable: %v", err)
	}
	mustTreeWrite(t, filepath.Join(f.lane, ".gitattributes"), "file filter=fixture\n")
	f.git(f.lane, "add", ".gitattributes")
	f.git(f.lane, "commit", "-qm", "activate filter")
	head := f.git(f.lane, "rev-parse", "HEAD")
	if err := f.s.checkoutFilters(f.ctx, f.repo, head); err == nil {
		t.Fatal("incoming tree filter was accepted")
	}
	assertAbsent(t, filepath.Join(f.repo, "filter-executed"))
	assertAbsent(t, filepath.Join(f.lane, "filter-executed"))
}
func TestCheckoutFilterProofIncludesLocalAndGlobalAttributes(t *testing.T) {
	for _, source := range []string{"local", "global"} {
		t.Run(source, func(t *testing.T) {
			f := newCleanupFixture(t)
			f.git(f.repo, "config", "filter.fixture.smudge", "touch filter-executed")
			attributes := filepath.Join(f.repo, ".git", "info", "attributes")
			if source == "global" {
				attributes = filepath.Join(f.base, "attributes")
				f.git(f.repo, "config", "core.attributesFile", attributes)
			}
			if err := os.MkdirAll(filepath.Dir(attributes), 0755); err != nil {
				t.Fatal(err)
			}
			mustTreeWrite(t, attributes, "file filter=fixture\n")
			if err := f.s.checkoutFilters(f.ctx, f.repo, f.head); err == nil {
				t.Fatal("effective external attributes were accepted")
			}
			assertAbsent(t, filepath.Join(f.repo, "filter-executed"))
		})
	}
}

func TestUnusedFilterAllowsCleanupButIncomingFilterBlocksSync(t *testing.T) {
	t.Run("unused configured driver", func(t *testing.T) {
		f := newCleanupFixture(t)
		f.git(f.repo, "config", "filter.fixture.smudge", "touch filter-executed")
		assertCleanupState(t, f.execute(), "REMOVED")
		assertAbsent(t, filepath.Join(f.repo, "filter-executed"))
	})
	t.Run("incoming active driver", func(t *testing.T) {
		f := newCleanupFixture(t)
		f.git(f.repo, "config", "filter.fixture.smudge", "touch filter-executed")
		mustTreeWrite(t, filepath.Join(f.lane, ".gitattributes"), "file filter=fixture\n")
		f.git(f.lane, "add", ".gitattributes")
		f.git(f.lane, "commit", "-qm", "incoming filter")
		f.head = f.git(f.lane, "rev-parse", "HEAD")
		f.git(f.lane, "push", "-q", "origin", "HEAD:main")
		f.receipt = f.record("primary", f.repo, "")
		f.options.FinishedHead = f.head
		assertCleanupState(t, f.execute(), "KEEP_LOCAL_STATE")
		assertAbsent(t, filepath.Join(f.repo, "filter-executed"))
		if f.git(f.repo, "rev-parse", "HEAD") == f.head {
			t.Fatal("unsafe checkout executed")
		}
	})
}

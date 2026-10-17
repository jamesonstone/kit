package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMintWriterLockSerializesAndReleases(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/mint.yaml"
	writePipelineFixture(t, root, path, "old\n")
	change := *newMintPipelineFileChange(root, path, "old\n", "new\n", instructionFileUpdated)
	if err := withMintWorkflowLock(change.absolutePath, func() error {
		if err := applyMintWorkflowChange(change); err == nil {
			t.Fatal("second writer acquired held lock")
		}
		if err := rollbackMintWorkflowChange(change); err == nil {
			t.Fatal("rollback acquired held lock")
		}
		if got := readFile(t, change.absolutePath); got != "old\n" {
			t.Fatal(got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := applyMintWorkflowChange(change); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(change.absolutePath + ".kit-mint.lock"); !os.IsNotExist(err) {
		t.Fatal("lock left behind")
	}
}
func TestMintWriterDoesNotReclaimExistingLock(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/mint.yaml"
	writePipelineFixture(t, root, path, "old\n")
	change := *newMintPipelineFileChange(root, path, "old\n", "new\n", instructionFileUpdated)
	lockPath := change.absolutePath + ".kit-mint.lock"
	writePipelineFixture(t, root, path+".kit-mint.lock", "another writer")
	if err := applyMintWorkflowChange(change); err == nil || !strings.Contains(err.Error(), lockPath) {
		t.Fatal(err)
	}
	if got := readFile(t, lockPath); got != "another writer" {
		t.Fatal("existing lock reclaimed")
	}
	if got := readFile(t, change.absolutePath); got != "old\n" {
		t.Fatal(got)
	}
}
func TestMintRollbackRetainsCreatedWorkflowAndRollsBackOtherFiles(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/mint.yaml"
	first := *newInitRefreshFileChange(root, "first.txt", "", "batch", instructionFileCreated)
	workflow := *newMintPipelineFileChange(root, path, "", "complete\n", instructionFileCreated)
	later := *newInitRefreshFileChange(root, "later.txt", "", "batch", instructionFileCreated)
	writer := func(change initRefreshFileChange) error {
		if change.relativePath == later.relativePath {
			return errors.New("later write failed")
		}
		return applyInitRefreshFileChange(change)
	}
	err := applyInitRefreshFileChangesWithWriter([]initRefreshFileChange{first, workflow, later}, rollbackInitRefreshFileChange, writer)
	if err == nil || !strings.Contains(err.Error(), workflow.absolutePath) || !strings.Contains(err.Error(), "retained") {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(root, path)); got != "complete\n" {
		t.Fatal(got)
	}
	if _, err := os.Stat(first.absolutePath); !os.IsNotExist(err) {
		t.Fatal("other file not rolled back")
	}
}

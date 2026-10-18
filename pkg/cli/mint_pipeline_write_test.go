package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMintWriteRejectsEditDuringEarlierBatchWrite(t *testing.T) {
	for _, created := range []bool{false, true} {
		root := t.TempDir()
		path := ".github/workflows/mint.yaml"
		before := "old\n"
		result := instructionFileUpdated
		if created {
			before = ""
			result = instructionFileCreated
		} else {
			writePipelineFixture(t, root, path, before)
		}
		workflow := *newMintPipelineFileChange(root, path, before, "planned\n", result)
		first := *newInitRefreshFileChange(root, "first.txt", "", "batch", instructionFileCreated)
		concurrent := "human edit\n"
		writer := func(change initRefreshFileChange) error {
			if err := applyInitRefreshFileChange(change); err != nil {
				return err
			}
			if change.relativePath == first.relativePath {
				writePipelineFixture(t, root, path, concurrent)
			}
			return nil
		}
		err := applyInitRefreshFileChangesWithWriter([]initRefreshFileChange{first, workflow}, rollbackInitRefreshFileChange, writer)
		if err == nil {
			t.Fatal("changed workflow replaced")
		}
		if got := readFile(t, filepath.Join(root, path)); got != concurrent {
			t.Fatal(got)
		}
		if _, err := os.Stat(first.absolutePath); !os.IsNotExist(err) {
			t.Fatal("earlier batch write not rolled back")
		}
	}
}
func TestMintRollbackPreservesConcurrentEdit(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/mint.yaml"
	writePipelineFixture(t, root, path, "old\n")
	workflow := *newMintPipelineFileChange(root, path, "old\n", "planned\n", instructionFileUpdated)
	later := *newInitRefreshFileChange(root, "later.txt", "", "batch", instructionFileCreated)
	writer := func(change initRefreshFileChange) error {
		if change.relativePath == later.relativePath {
			writePipelineFixture(t, root, path, "human edit\n")
			return errors.New("later write failed")
		}
		return applyInitRefreshFileChange(change)
	}
	err := applyInitRefreshFileChangesWithWriter([]initRefreshFileChange{workflow, later}, rollbackInitRefreshFileChange, writer)
	if err == nil {
		t.Fatal("failure ignored")
	}
	if got := readFile(t, filepath.Join(root, path)); got != "human edit\n" {
		t.Fatal("rollback overwrote human edit")
	}
}
func TestMintConditionalWritePreservesMode(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/mint.yaml"
	writePipelineFixture(t, root, path, "old\n")
	absolute := filepath.Join(root, path)
	if err := os.Chmod(absolute, 0600); err != nil {
		t.Fatal(err)
	}
	change := newMintPipelineFileChange(root, path, "old\n", "planned\n", instructionFileUpdated)
	if err := applyMintWorkflowChange(*change); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(absolute)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("%v %v", info, err)
	}
}

func TestMintExclusiveCreateFailureRollsBackEarlierFiles(t *testing.T) {
	for _, closeFailure := range []bool{false, true} {
		root := t.TempDir()
		path := ".github/workflows/mint.yaml"
		workflow := *newMintPipelineFileChange(root, path, "", "planned\n", instructionFileCreated)
		first := *newInitRefreshFileChange(root, "first.txt", "", "batch", instructionFileCreated)
		writer := func(change initRefreshFileChange) error {
			if !change.strictMintWorkflow {
				return applyInitRefreshFileChange(change)
			}
			return createMintWorkflowExclusive(change.absolutePath, change.after, func(file *os.File, content string) error {
				if closeFailure {
					return file.Close()
				}
				if _, err := file.WriteString("partial"); err != nil {
					return err
				}
				return errors.New("write failed")
			})
		}
		if err := applyInitRefreshFileChangesWithWriter([]initRefreshFileChange{first, workflow}, rollbackInitRefreshFileChange, writer); err == nil {
			t.Fatal("failure ignored")
		}
		for _, path := range []string{first.absolutePath, workflow.absolutePath} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatal("partial batch destination exists")
			}
		}
		temps, err := filepath.Glob(filepath.Join(root, ".github/workflows/.kit-mint-*"))
		if err != nil || len(temps) != 0 {
			t.Fatalf("%v %v", temps, err)
		}
	}
}
func TestMintExclusiveCreatePreservesFileAppearingDuringStage(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".github/workflows/mint.yaml")
	err := createMintWorkflowExclusive(path, "planned\n", func(file *os.File, content string) error {
		if err := os.WriteFile(path, []byte("human edit\n"), 0600); err != nil {
			return err
		}
		_, err := file.WriteString(content)
		return err
	})
	if err == nil {
		t.Fatal("concurrent destination replaced")
	}
	if got := readFile(t, path); got != "human edit\n" {
		t.Fatal(got)
	}
}

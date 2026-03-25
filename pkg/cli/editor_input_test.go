package cli

import (
	"os/exec"
	stdreflect "reflect"
	"testing"
)

func TestResolveEditorCommand_UsesPreferredVimEditorForVimFlag(t *testing.T) {
	restore := stubLookPath(map[string]string{
		"nvim": "/usr/local/bin/nvim",
		"vim":  "/usr/bin/vim",
	})
	defer restore()

	got, err := newFreeTextInputConfig(true, "").resolveEditorCommand()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{"/usr/local/bin/nvim"}
	if !stdreflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestResolveEditorCommand_UsesVimAliasFromEditorFlag(t *testing.T) {
	restore := stubLookPath(map[string]string{
		"vim": "/usr/bin/vim",
	})
	defer restore()

	got, err := newFreeTextInputConfig(false, "vim").resolveEditorCommand()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{"/usr/bin/vim"}
	if !stdreflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestResolveEditorCommand_ExplicitEditorOverridesVimFlag(t *testing.T) {
	restore := stubLookPath(map[string]string{
		"nvim": "/custom/nvim",
	})
	defer restore()

	got, err := newFreeTextInputConfig(true, "nvim").resolveEditorCommand()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{"/custom/nvim"}
	if !stdreflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestResolveEditorCommand_ErrorsWhenNoVimEditorExists(t *testing.T) {
	restore := stubLookPath(map[string]string{})
	defer restore()

	if _, err := newFreeTextInputConfig(true, "").resolveEditorCommand(); err == nil {
		t.Fatalf("expected error when no vim-compatible editor is available")
	}
}

func TestFinalizeEditorInput_NormalizesAndDetectsChange(t *testing.T) {
	got, changed, err := finalizeEditorInput("", []byte("  first line\nsecond line  \n"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !changed {
		t.Fatalf("expected edited content to be marked changed")
	}
	if got != "first line\nsecond line" {
		t.Fatalf("expected normalized content, got %q", got)
	}
}

func TestFinalizeEditorInput_UnchangedContentReturnsFalse(t *testing.T) {
	got, changed, err := finalizeEditorInput("", []byte(""))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if changed {
		t.Fatalf("expected unchanged content to remain unchanged")
	}
	if got != "" {
		t.Fatalf("expected empty content, got %q", got)
	}
}

func stubLookPath(entries map[string]string) func() {
	previous := lookPath
	lookPath = func(file string) (string, error) {
		if path, ok := entries[file]; ok {
			return path, nil
		}
		return "", &exec.Error{Name: file, Err: exec.ErrNotFound}
	}

	return func() {
		lookPath = previous
	}
}

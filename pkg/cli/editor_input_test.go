package cli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	stdreflect "reflect"
	"strings"
	"testing"
)

func TestResolveEditorCommand_UsesPreferredVimEditorForVimFlag(t *testing.T) {
	restore := stubLookPath(map[string]string{
		"nvim": "/usr/local/bin/nvim",
		"vim":  "/usr/bin/vim",
	})
	defer restore()

	got, err := newFreeTextInputConfig(true, "", false, false).resolveEditorCommand()
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

	got, err := newFreeTextInputConfig(false, "vim", false, false).resolveEditorCommand()
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

	got, err := newFreeTextInputConfig(true, "nvim", false, false).resolveEditorCommand()
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

	if _, err := newFreeTextInputConfig(true, "", false, false).resolveEditorCommand(); err == nil {
		t.Fatalf("expected error when no vim-compatible editor is available")
	}
}

func TestResolveEditorCommand_UsesDefaultEditorWhenConfigured(t *testing.T) {
	restore := stubLookPath(map[string]string{
		"nvim": "/usr/local/bin/nvim",
	})
	defer restore()

	got, err := newFreeTextInputConfig(false, "", false, true).resolveEditorCommand()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{"/usr/local/bin/nvim"}
	if !stdreflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestFreeTextInputConfig_InlineOverridesDefaultEditor(t *testing.T) {
	if newFreeTextInputConfig(false, "", true, true).usesEditor() {
		t.Fatal("expected --inline to disable editor mode")
	}
}

func TestInlineFlagRegisteredOnlyOnInlineCapableCommands(t *testing.T) {
	brainstorm, _, err := rootCmd.Find([]string{"brainstorm"})
	if err != nil {
		t.Fatalf("rootCmd.Find(brainstorm) error = %v", err)
	}
	if brainstorm.Flags().Lookup("inline") == nil {
		t.Fatal("expected brainstorm to expose --inline")
	}

	spec, _, err := rootCmd.Find([]string{"spec"})
	if err != nil {
		t.Fatalf("rootCmd.Find(spec) error = %v", err)
	}
	if spec.Flags().Lookup("inline") == nil {
		t.Fatal("expected spec to expose --inline")
	}

	dispatch, _, err := rootCmd.Find([]string{"dispatch"})
	if err != nil {
		t.Fatalf("rootCmd.Find(dispatch) error = %v", err)
	}
	if dispatch.Flags().Lookup("inline") != nil {
		t.Fatal("expected dispatch not to expose --inline")
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

func TestPrintEditorLaunchInstructions(t *testing.T) {
	var output bytes.Buffer

	err := printEditorLaunchInstructions(
		&output,
		newFreeTextInputConfig(true, "", false, false),
		"dispatch tasks",
		"cancel",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rendered := output.String()
	checks := []string{
		"Step: dispatch tasks.",
		"vim-compatible editor",
		"Paste only the content for this response",
		"Quit without save to cancel",
		"Press any key to open the editor.",
	}

	for _, check := range checks {
		if !strings.Contains(rendered, check) {
			t.Fatalf("expected instructions to contain %q", check)
		}
	}
}

func TestReadEditorTextWaitsBeforeLaunchingEditor(t *testing.T) {
	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	defer func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	}()

	var sequence []string
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		sequence = append(sequence, "wait")
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, _ string, _ string) (string, bool, error) {
		sequence = append(sequence, "run")
		return "captured text", true, nil
	}

	got, err := readEditorText(
		newFreeTextInputConfig(true, "", false, false),
		"dispatch tasks",
		false,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != "captured text" {
		t.Fatalf("expected captured text, got %q", got)
	}

	want := []string{"wait", "run"}
	if !stdreflect.DeepEqual(sequence, want) {
		t.Fatalf("expected sequence %v, got %v", want, sequence)
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

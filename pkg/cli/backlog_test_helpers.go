package cli

import (
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func newBrainstormTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", false, "")
	cmd.Flags().Bool("prompt-only", false, "")
	return cmd
}

func stubBrainstormEditor(t *testing.T, thesis string) func() {
	t.Helper()

	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error { return nil }
	editorInputRunner = func(_ freeTextInputConfig, _ string, _ string) (string, bool, error) {
		return thesis, true, nil
	}

	return func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	}
}

func setBacklogFlagState(pickup, copy, outputOnly bool) func() {
	previousPickup := backlogPickup
	previousCopy := backlogCopy
	previousOutputOnly := backlogOutputOnly

	backlogPickup = pickup
	backlogCopy = copy
	backlogOutputOnly = outputOnly

	return func() {
		backlogPickup = previousPickup
		backlogCopy = previousCopy
		backlogOutputOnly = previousOutputOnly
	}
}

func setBrainstormFlagState(
	backlog bool,
	pickup bool,
	output string,
	copy bool,
	outputOnly bool,
	inline bool,
	useVim bool,
) func() {
	previousBacklog := brainstormBacklog
	previousPickup := brainstormPickup
	previousOutput := brainstormOutput
	previousCopy := brainstormCopy
	previousOutputOnly := brainstormOutputOnly
	previousInline := brainstormInline
	previousUseVim := brainstormUseVim
	previousEditor := brainstormEditor

	brainstormBacklog = backlog
	brainstormPickup = pickup
	brainstormOutput = output
	brainstormCopy = copy
	brainstormOutputOnly = outputOnly
	brainstormInline = inline
	brainstormUseVim = useVim
	brainstormEditor = ""

	return func() {
		brainstormBacklog = previousBacklog
		brainstormPickup = previousPickup
		brainstormOutput = previousOutput
		brainstormCopy = previousCopy
		brainstormOutputOnly = previousOutputOnly
		brainstormInline = previousInline
		brainstormUseVim = previousUseVim
		brainstormEditor = previousEditor
	}
}

func silenceStdout(t *testing.T) func() {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer

	return func() {
		_ = writer.Close()
		_, _ = io.Copy(io.Discard, reader)
		_ = reader.Close()
		os.Stdout = originalStdout
	}
}

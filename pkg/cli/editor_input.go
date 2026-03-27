package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	awaitEditorLaunchConfirmation = waitForEditorLaunchConfirmation
	editorInputRunner             = runEditorInput
	execCommand                   = exec.Command
	lookPath                      = exec.LookPath
)

type freeTextInputConfig struct {
	editor string
	useVim bool
}

func addFreeTextInputFlags(cmd *cobra.Command, useVim *bool, editor *string) {
	cmd.Flags().BoolVar(
		useVim,
		"vim",
		false,
		"open free-text prompts in a vim-compatible editor (shorthand for --editor=vim)",
	)
	cmd.Flags().StringVar(
		editor,
		"editor",
		"",
		"open free-text prompts in an editor; use 'vim' as an alias for --vim",
	)
}

func newFreeTextInputConfig(useVim bool, editor string) freeTextInputConfig {
	return freeTextInputConfig{
		editor: strings.TrimSpace(editor),
		useVim: useVim,
	}
}

func (c freeTextInputConfig) usesEditor() bool {
	return c.useVim || c.editor != ""
}

func (c freeTextInputConfig) editorLabel() string {
	if c.editor != "" && !strings.EqualFold(c.editor, "vim") {
		return c.editor
	}
	return "vim-compatible editor"
}

func (c freeTextInputConfig) resolveEditorCommand() ([]string, error) {
	switch {
	case strings.EqualFold(c.editor, "vim"):
		return resolveVimCommand()
	case c.editor != "":
		return resolveExactEditorCommand(c.editor)
	case c.useVim:
		return resolveVimCommand()
	default:
		return nil, fmt.Errorf("no editor configured")
	}
}

func resolveVimCommand() ([]string, error) {
	for _, candidate := range []string{"nvim", "vim", "vi"} {
		path, err := lookPath(candidate)
		if err == nil {
			return []string{path}, nil
		}
	}

	return nil, fmt.Errorf("no vim-compatible editor found (tried nvim, vim, vi)")
}

func resolveExactEditorCommand(editor string) ([]string, error) {
	fields := strings.Fields(editor)
	if len(fields) == 0 {
		return nil, fmt.Errorf("editor cannot be empty")
	}

	path, err := lookPath(fields[0])
	if err != nil {
		return nil, fmt.Errorf("editor %q not found", fields[0])
	}

	fields[0] = path
	return fields, nil
}

func readEditorText(inputCfg freeTextInputConfig, fieldName string, emptyAllowed bool) (string, error) {
	cancelAction := "cancel"
	if emptyAllowed {
		cancelAction = "skip"
	}
	if err := printEditorLaunchInstructions(os.Stdout, inputCfg, fieldName, cancelAction); err != nil {
		return "", err
	}
	if err := awaitEditorLaunchConfirmation(os.Stdin, os.Stdout); err != nil {
		return "", err
	}

	text, changed, err := editorInputRunner(inputCfg, fieldName, "")
	if err != nil {
		return "", err
	}

	if !changed {
		if emptyAllowed {
			return "", nil
		}
		return "", fmt.Errorf("%s entry cancelled", fieldName)
	}

	if text == "" {
		if emptyAllowed {
			return "", nil
		}
		return "", fmt.Errorf("%s cannot be empty", fieldName)
	}

	return text, nil
}

func printEditorLaunchInstructions(
	output io.Writer,
	inputCfg freeTextInputConfig,
	fieldName string,
	cancelAction string,
) error {
	_, err := fmt.Fprintf(
		output,
		dim+"Step: %s. Paste only the content for this response into the %s that opens next. Save and quit to submit. Quit without save to %s.\nPress any key to open the editor."+reset+"\n",
		fieldName,
		inputCfg.editorLabel(),
		cancelAction,
	)
	return err
}

func waitForEditorLaunchConfirmation(input *os.File, output io.Writer) error {
	if input == nil {
		return nil
	}

	fd := int(input.Fd())
	if !term.IsTerminal(fd) {
		return nil
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to enter raw mode for editor launch: %w", err)
	}
	defer term.Restore(fd, state)

	var key [1]byte
	if _, err := input.Read(key[:]); err != nil {
		return fmt.Errorf("failed to read key press before opening editor: %w", err)
	}

	if key[0] == 3 {
		return fmt.Errorf("editor launch cancelled")
	}

	if output != nil {
		if _, err := io.WriteString(output, "\n"); err != nil {
			return fmt.Errorf("failed to write editor launch newline: %w", err)
		}
	}

	return nil
}

func runEditorInput(inputCfg freeTextInputConfig, fieldName, initialContent string) (string, bool, error) {
	editorCommand, err := inputCfg.resolveEditorCommand()
	if err != nil {
		return "", false, err
	}

	tempFile, err := os.CreateTemp("", "kit-input-*.md")
	if err != nil {
		return "", false, fmt.Errorf("failed to create temp file for %s: %w", fieldName, err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.WriteString(initialContent); err != nil {
		tempFile.Close()
		return "", false, fmt.Errorf("failed to seed temp file for %s: %w", fieldName, err)
	}
	if err := tempFile.Close(); err != nil {
		return "", false, fmt.Errorf("failed to close temp file for %s: %w", fieldName, err)
	}

	args := append(editorCommand[1:], tempPath)
	cmd := execCommand(editorCommand[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", false, fmt.Errorf("failed to open editor for %s: %w", fieldName, err)
	}

	edited, err := os.ReadFile(tempPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read editor output for %s: %w", fieldName, err)
	}

	return finalizeEditorInput(initialContent, edited)
}

func finalizeEditorInput(initialContent string, edited []byte) (string, bool, error) {
	raw := string(edited)
	return normalizeSpecAnswer(raw), raw != initialContent, nil
}

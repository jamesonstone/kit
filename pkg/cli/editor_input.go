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
	defaultEditor bool
	editor        string
	inline        bool
	useVim        bool
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

func addInlineTextInputFlag(cmd *cobra.Command, inline *bool) {
	cmd.Flags().BoolVar(
		inline,
		"inline",
		false,
		"use inline multiline prompts instead of opening a vim-compatible editor",
	)
}

func newFreeTextInputConfig(
	useVim bool,
	editor string,
	inline bool,
	defaultEditor bool,
) freeTextInputConfig {
	return freeTextInputConfig{
		defaultEditor: defaultEditor,
		editor:        strings.TrimSpace(editor),
		inline:        inline,
		useVim:        useVim,
	}
}

func (c freeTextInputConfig) usesEditor() bool {
	if c.inline {
		return false
	}

	return c.useVim || c.editor != "" || c.defaultEditor
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
	case c.useVim || c.defaultEditor:
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
	style := styleForWriter(output)

	if divider := style.sectionDivider(); divider != "" {
		if _, err := fmt.Fprintln(output, divider); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(output, style.title("📝", fmt.Sprintf("Step: %s.", fieldName))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(
		output,
		"%s\n",
		style.muted(fmt.Sprintf(
			"Paste only the content for this response into the %s that opens next.",
			inputCfg.editorLabel(),
		)),
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(
		output,
		"%s\n",
		style.muted(fmt.Sprintf("Save and quit to submit. Quit without save to %s.", cancelAction)),
	); err != nil {
		return err
	}
	_, err := fmt.Fprintf(output, "%s\n", style.label("Press any key to open the editor."))
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

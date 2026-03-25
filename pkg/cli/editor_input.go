package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	execCommand = exec.Command
	lookPath    = exec.LookPath
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

	fmt.Printf(
		dim+"Opening %s for %s. Save and quit to submit. Quit without save to %s."+reset+"\n",
		inputCfg.editorLabel(),
		fieldName,
		cancelAction,
	)

	text, changed, err := runEditorInput(inputCfg, fieldName, "")
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

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
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
		"open free-text prompts in a vim-compatible editor instead of the default editor",
	)
	cmd.Flags().StringVar(
		editor,
		"editor",
		"",
		"open free-text prompts in a specific editor command; defaults to $EDITOR when omitted",
	)
}

func addInlineTextInputFlag(cmd *cobra.Command, inline *bool) {
	cmd.Flags().BoolVar(
		inline,
		"inline",
		false,
		"use inline multiline prompts instead of opening the default editor",
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
	if strings.EqualFold(c.editor, "vim") || c.useVim {
		return "vim-compatible editor"
	}
	if c.defaultEditor {
		if editor, ok := defaultEditorEnv(); ok {
			return "$EDITOR (" + editor + ")"
		}
		return "default editor"
	}
	return "editor"
}

func (c freeTextInputConfig) resolveEditorCommand() ([]string, error) {
	switch {
	case strings.EqualFold(c.editor, "vim"):
		return resolveVimCommand()
	case c.editor != "":
		return resolveExactEditorCommand(c.editor)
	case c.useVim:
		return resolveVimCommand()
	case c.defaultEditor:
		return resolveDefaultEditorCommand()
	default:
		return nil, fmt.Errorf("no editor configured")
	}
}

func resolveDefaultEditorCommand() ([]string, error) {
	if editor, ok := defaultEditorEnv(); ok {
		return resolveExactEditorCommand(editor)
	}
	return resolveVimCommand()
}

func defaultEditorEnv() (string, bool) {
	editor, ok := os.LookupEnv("EDITOR")
	editor = strings.TrimSpace(editor)
	return editor, ok && editor != ""
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
	return readEditorTextWithInitialContent(inputCfg, fieldName, "", emptyAllowed, true)
}

func readEditorTextWithInitialContent(
	inputCfg freeTextInputConfig,
	fieldName string,
	initialContent string,
	emptyAllowed bool,
	requireChange bool,
) (string, error) {
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

	text, changed, err := editorInputRunner(inputCfg, fieldName, initialContent)
	if err != nil {
		return "", err
	}

	if !changed && requireChange {
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

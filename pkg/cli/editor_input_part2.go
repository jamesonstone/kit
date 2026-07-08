package cli

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

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
	defer func() { _ = term.Restore(fd, state) }()

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
	defer func() { _ = os.Remove(tempPath) }()

	if _, err := tempFile.WriteString(initialContent); err != nil {
		_ = tempFile.Close()
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

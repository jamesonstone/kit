package cli

import (
	"fmt"
	"os"
)

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

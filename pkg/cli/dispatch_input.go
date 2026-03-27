package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type dispatchInputSource string

const (
	dispatchInputSourceFile   dispatchInputSource = "file"
	dispatchInputSourceStdin  dispatchInputSource = "stdin"
	dispatchInputSourceEditor dispatchInputSource = "editor"
)

func resolveDispatchInputSource(filePath string, stdinIsTerminal bool) dispatchInputSource {
	if strings.TrimSpace(filePath) != "" {
		return dispatchInputSourceFile
	}
	if !stdinIsTerminal {
		return dispatchInputSourceStdin
	}

	return dispatchInputSourceEditor
}

func loadDispatchInput(
	filePath string,
	inputCfg freeTextInputConfig,
) (string, dispatchInputSource, error) {
	inputSource := resolveDispatchInputSource(filePath, isTerminal())

	switch inputSource {
	case dispatchInputSourceFile:
		content, err := readDispatchFile(filePath)
		return content, inputSource, err
	case dispatchInputSourceStdin:
		content, err := readDispatchStdin()
		return content, inputSource, err
	default:
		content, err := readDispatchEditor(inputCfg)
		return content, inputSource, err
	}
}

func readDispatchFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	return normalizeDispatchRawInput(string(content)), nil
}

func readDispatchStdin() (string, error) {
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read task input from stdin: %w", err)
	}

	return normalizeDispatchRawInput(string(content)), nil
}

func readDispatchEditor(inputCfg freeTextInputConfig) (string, error) {
	return readEditorText(inputCfg, "dispatch tasks", false)
}

func normalizeDispatchRawInput(raw string) string {
	replacedCRLF := strings.ReplaceAll(raw, "\r\n", "\n")
	replacedCR := strings.ReplaceAll(replacedCRLF, "\r", "\n")
	return normalizeSpecAnswer(replacedCR)
}

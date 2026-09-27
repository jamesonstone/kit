package cli

import (
	"fmt"
	"strings"
)

var clipboardCopyFunc = copyToClipboard

func formatAgentInstructionBlock(prompt string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(prompt)
	if !strings.HasSuffix(prompt, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("---\n")
	return sb.String()
}

func outputPromptWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	return writePromptWithClipboardDefault(prompt, outputOnly, copy)
}

func writePromptWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	shouldCopy := !outputOnly || copy
	if shouldCopy {
		if err := clipboardCopyFunc(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
	}

	if outputOnly {
		fmt.Print(prompt)
		return nil
	}

	fmt.Println(styleForStdout().clipboardAcknowledgement())
	return nil
}

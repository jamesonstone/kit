package cli

import (
	"fmt"
)

var clipboardCopyFunc = copyToClipboard

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

package cli

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var clipboardCopyFunc = copyToClipboard

// stdoutIsTerminal decides the default prompt destination: people at a
// terminal get the clipboard, while agents and scripts get stdout.
var stdoutIsTerminal = func() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func outputPromptWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	return writePromptWithClipboardDefault(prompt, outputOnly, copy)
}

func writePromptWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	shouldCopy := promptCopiedByDefault(outputOnly, copy)
	if !shouldCopy {
		outputOnly = true
	}
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

// promptCopiedByDefault reports whether a prompt goes to the clipboard: always
// with --copy, otherwise only at a terminal without --output-only.
func promptCopiedByDefault(outputOnly, copy bool) bool {
	return copy || (!outputOnly && stdoutIsTerminal())
}

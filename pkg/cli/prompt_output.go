package cli

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var clipboardCopyFunc = copyToClipboard

// stdoutIsTerminal decides the default prompt destination: people at a
// terminal get the clipboard, while agents and scripts get stdout.
var stdoutIsTerminal = func() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

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
	if !outputOnly && !copy && !stdoutIsTerminal() {
		outputOnly = true
	}
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

package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func writePROrchestratePrompt(cmd *cobra.Command, prompt string, interactive, outputOnly, copyRequested bool) error {
	rawOutput := outputOnly || !interactive
	shouldCopy := copyRequested || (interactive && !outputOnly)
	if shouldCopy {
		if err := clipboardCopyFunc(prompt); err != nil {
			return fmt.Errorf("failed to copy release orchestration prompt to clipboard: %w", err)
		}
	}
	if rawOutput {
		_, err := io.WriteString(cmd.OutOrStdout(), prompt)
		return err
	}
	_, err := fmt.Fprintln(cmd.OutOrStdout(), styleForWriter(cmd.OutOrStdout()).clipboardAcknowledgement())
	return err
}

package agentcli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

type exitError struct {
	code int
	err  error
}

func (err *exitError) Error() string { return err.err.Error() }
func (err *exitError) Unwrap() error { return err.err }

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "kit",
		Short:         "Repository-native contracts for coding agents",
		Long:          "Kit materializes registry-backed rules and workflows, then resolves the deterministic repository contract coding agents follow.",
		Version:       Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(newInitCommand())
	root.AddCommand(newReconcileCommand())
	root.AddCommand(newRulesCommand())
	root.AddCommand(newRegistryCommand())
	root.AddCommand(newContractCommand())
	root.AddCommand(newPRCommand())
	return root
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		var codeErr *exitError
		if errors.As(err, &codeErr) {
			fmt.Fprintln(os.Stderr, codeErr.Error())
			os.Exit(codeErr.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeJSON(writer io.Writer, value interface{}) error {
	encoder := newJSONEncoder(writer)
	return encoder.Encode(value)
}

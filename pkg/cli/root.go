package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "kit",
	Short: "🧰 Kit v2 is a general-purpose harness for thought work",
	Long: banner() + `
Kit v2 is a general-purpose harness for disciplined thought work.
Its strongest engine is a document-first, spec-driven workflow, but the
harness also supports ad hoc execution, catch-up, handoff, summarization,
review, and orchestration.

The current command surface is packaged around repository and software
workflows, but the underlying harness patterns generalize to research,
strategy, operations, writing, policy, and other structured fields.

` + flowDiagram(),
	Version: Version,
}

func init() {
	rootCmd.SetVersionTemplate("kit version {{.Version}}\n")
	configureRootHelp()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		var exitErr *cliExitError
		if errors.As(err, &exitErr) {
			if !exitErr.silent {
				fmt.Fprintln(os.Stderr, exitErr.Error())
			}
			os.Exit(exitErr.code)
		}
		var silentErr *silentCLIError
		if !errors.As(err, &silentErr) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

type silentCLIError struct {
	err error
}

func (e *silentCLIError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *silentCLIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

type cliExitError struct {
	err    error
	code   int
	silent bool
}

func newCLIExitError(err error, code int, silent bool) *cliExitError {
	if code == 0 {
		code = 1
	}
	return &cliExitError{err: err, code: code, silent: silent}
}

func (e *cliExitError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *cliExitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jamesonstone/kit/v3/internal/commandset"
	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/usage"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:               "kit",
	Short:             "🧰 Kit resolves repository-local evidence for coding agents",
	Long:              rootLong(humanOutputStyle{}),
	Version:           Version,
	PersistentPreRunE: runAutomaticConfigCheck,
}

func init() {
	rootCmd.SetVersionTemplate("kit version {{.Version}}\n")
	configureRootHelp()
}

func Execute() {
	pruneCommandTree(rootCmd, "")
	started := time.Now()
	executed, err := rootCmd.ExecuteC()
	recordUsage(executed, err, time.Since(started))
	if err != nil {
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

func pruneCommandTree(parent *cobra.Command, parentPath string) {
	for _, child := range append([]*cobra.Command(nil), parent.Commands()...) {
		path := strings.TrimSpace(parentPath + " " + child.Name())
		if !commandset.IsProtectedOrParent(path) {
			parent.RemoveCommand(child)
			continue
		}
		if path == "completion" {
			continue
		}
		pruneCommandTree(child, path)
	}
}

func recordUsage(executed *cobra.Command, commandErr error, elapsed time.Duration) {
	if executed == nil {
		return
	}
	path := strings.TrimPrefix(executed.CommandPath(), rootCmd.Name()+" ")
	if strings.HasPrefix(path, "completion ") {
		path = "completion"
	}
	if !commandset.IsTelemetryPath(path) {
		return
	}
	exitCode := 0
	if commandErr != nil {
		exitCode = 1
		var exitErr *cliExitError
		if errors.As(commandErr, &exitErr) {
			exitCode = exitErr.code
		}
	}
	projectRoot, found, _ := config.FindProjectRootOptional()
	if !found {
		projectRoot = ""
	}
	_ = usage.Record(usage.RecordInput{
		Command: path, Version: Version, ExitCode: exitCode, Elapsed: elapsed,
		ProjectRoot: projectRoot,
		Interactive: term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())),
	})
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

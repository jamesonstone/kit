package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
	contextcontract "github.com/jamesonstone/kit/v3/internal/context"
)

type contextResolveOptions struct {
	workflow   string
	feature    string
	paths      []string
	jsonOutput bool
}

func init() {
	rootCmd.AddCommand(newContextCommand())
}

func newContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Resolve deterministic repository-local coding-agent context",
	}
	cmd.AddCommand(newContextResolveCommand())
	return cmd
}

func newContextResolveCommand() *cobra.Command {
	opts := &contextResolveOptions{workflow: "implementation-delivery"}
	cmd := &cobra.Command{
		Use:           "resolve",
		Short:         "Resolve ordered local workflow, rule, specification, and reference evidence",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runContextResolve(cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.workflow, "workflow", opts.workflow, "workflow slug to resolve")
	cmd.Flags().StringVar(&opts.feature, "feature", "", "feature slug or directory to include")
	cmd.Flags().StringArrayVar(&opts.paths, "path", nil, "required repository-relative path hint; repeatable")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "emit versioned machine-readable JSON")
	return cmd
}

func runContextResolve(cmd *cobra.Command, opts *contextResolveOptions) error {
	projectRoot, found, err := config.FindProjectRootOptional()
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("kit project not initialized: run `kit init` before `kit context resolve`")
	}
	contract := contextcontract.Resolve(projectRoot, contextcontract.Request{
		Workflow: opts.workflow,
		Feature:  opts.feature,
		Paths:    opts.paths,
	})
	if opts.jsonOutput {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(contract); err != nil {
			return err
		}
	} else if err := renderContextContract(cmd, contract); err != nil {
		return err
	}
	if contract.Blocked {
		return newCLIExitError(errors.New("context resolution blocked by required local evidence"), 2, true)
	}
	return nil
}

package agentcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

type reconcileOptions struct {
	apply   bool
	diff    bool
	json    bool
	accepts []string
}

func newReconcileCommand() *cobra.Command {
	options := &reconcileOptions{}
	command := &cobra.Command{
		Use:   "reconcile",
		Short: "Preview or apply coding-agent contract drift",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runReconcile(command, *options)
		},
	}
	command.Flags().BoolVar(&options.apply, "apply", false, "apply conflict-free planned changes")
	command.Flags().BoolVar(&options.diff, "diff", false, "print a unified change preview")
	command.Flags().BoolVar(&options.json, "json", false, "emit the reconciliation plan as JSON")
	command.Flags().StringArrayVar(&options.accepts, "accept-registry", nil, "accept registry content for one exact kind/slug; repeat as needed")
	return command
}

func runReconcile(command *cobra.Command, options reconcileOptions) error {
	root, err := registry.FindProjectRoot(mustGetwd())
	if err != nil {
		return err
	}
	cfg, _, err := registry.LoadProject(root)
	if err != nil {
		return err
	}
	accepts, err := parseAccepts(options.accepts)
	if err != nil {
		return err
	}
	plan, err := registry.BuildReconcilePlan(command.Context(), root, sourceFor(root, cfg.Registry.Source), accepts)
	if err != nil {
		return err
	}
	applied := false
	if options.apply {
		if err := registry.ApplyPlan(root, plan); err != nil {
			return err
		}
		applied = true
	}
	if options.json {
		diff := ""
		if options.diff {
			diff = registry.RenderDiff(plan)
		}
		err = writeJSON(command.OutOrStdout(), struct {
			Applied bool          `json:"applied"`
			Plan    registry.Plan `json:"plan"`
			Diff    string        `json:"diff,omitempty"`
		}{Applied: applied, Plan: plan, Diff: diff})
	} else {
		if options.diff {
			if _, writeErr := fmt.Fprint(command.OutOrStdout(), registry.RenderDiff(plan)); writeErr != nil {
				return writeErr
			}
		}
		err = writePlanHuman(command.OutOrStdout(), plan, applied)
	}
	if err != nil {
		return err
	}
	if plan.State == "attention-needed" {
		return &exitError{code: 2, err: fmt.Errorf("reconciliation requires explicit conflict resolution")}
	}
	return nil
}

func parseAccepts(values []string) (map[string]bool, error) {
	result := map[string]bool{}
	for _, value := range values {
		parts := strings.Split(value, "/")
		if len(parts) != 2 || (parts[0] != registry.KindRuleset && parts[0] != registry.KindWorkflow) || !registry.ValidSlug(parts[1]) {
			return nil, fmt.Errorf("--accept-registry value %q must be ruleset/<slug> or workflow/<slug>", value)
		}
		result[value] = true
	}
	return result, nil
}

func mustGetwd() string {
	path, err := os.Getwd()
	if err != nil {
		return "."
	}
	return path
}

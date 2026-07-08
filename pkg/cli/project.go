package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

type projectRefreshOptions struct {
	Now              bool
	ConstitutionOnly bool
	OutputOnly       bool
	Copy             bool
}

func init() {
	rootCmd.AddCommand(newProjectCommand())
}

func newProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Run project-level Kit maintenance workflows",
		Long: `Run project-level Kit maintenance workflows.

Use kit project refresh to generate a semantic refresh prompt for durable
project-level documentation, especially docs/CONSTITUTION.md. This is separate
from kit reconcile, which refreshes structural Kit-managed files.`,
	}
	cmd.AddCommand(newProjectRefreshCommand())
	return cmd
}

func newProjectRefreshCommand() *cobra.Command {
	opts := projectRefreshOptions{}
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Generate or record a semantic project documentation refresh",
		Long: `Generate or record a semantic project documentation refresh.

By default this emits the same project-refresh prompt as the prompt library
entry and uses the shared clipboard-first prompt behavior. The prompt asks an
agent to inspect completed work, identify durable project-level rules,
constraints, vocabulary, conventions, and workflow changes, and update
docs/CONSTITUTION.md only after review.

Use --now after the reviewed semantic refresh is complete to update the tracked
cadence state in .kit.yaml. This command never rewrites docs/CONSTITUTION.md
automatically.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectRefreshCommand(cmd, opts, time.Now().UTC())
		},
	}
	cmd.Flags().BoolVar(&opts.Now, "now", false, "mark the reviewed Constitution refresh complete in .kit.yaml")
	cmd.Flags().BoolVar(&opts.ConstitutionOnly, "constitution-only", false, "generate a prompt scoped to docs/CONSTITUTION.md only")
	cmd.Flags().BoolVar(&opts.OutputOnly, "output-only", false, "print the prompt instead of copying it")
	cmd.Flags().BoolVar(&opts.Copy, "copy", false, "copy prompt even with --output-only")
	return cmd
}

func runProjectRefreshCommand(cmd *cobra.Command, opts projectRefreshOptions, now time.Time) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if opts.Now {
		if opts.Copy {
			return fmt.Errorf("--copy cannot be used with --now")
		}
		status, err := recordProjectRefreshReview(projectRoot, cfg, now)
		if err != nil {
			return err
		}
		if opts.OutputOnly {
			return printProjectRefreshStatusSummary(cmd.OutOrStdout(), status)
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "✅ Project refresh review recorded."); err != nil {
			return err
		}
		return printProjectRefreshStatusSummary(cmd.OutOrStdout(), status)
	}

	status, err := calculateProjectRefreshStatus(projectRoot, cfg, now)
	if err != nil {
		return err
	}
	prompt := buildProjectRefreshPromptWithOptions(projectRoot, cfg, projectRefreshPromptOptions{
		ConstitutionOnly: opts.ConstitutionOnly,
		Status:           status,
	})
	if !opts.OutputOnly {
		printWorkflowInstructions("project refresh", []string{
			"review the generated prompt before applying semantic documentation changes",
			"run `kit project refresh --now` only after the Constitution refresh is complete",
		})
	}
	return outputPromptWithClipboardDefault(prompt, opts.OutputOnly, opts.Copy)
}

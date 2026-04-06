package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

var (
	backlogCopy       bool
	backlogOutputOnly bool
	backlogPickup     bool
)

var backlogCmd = &cobra.Command{
	Use:   "backlog [feature]",
	Short: "List deferred backlog items or pick one up",
	Long: `List deferred backlog items captured as paused brainstorm-phase features.

Use --pickup to resume one of those items and output the standard brainstorm
planning prompt. Use ` + "`kit resume`" + ` as the canonical general resume flow.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBacklog,
}

func init() {
	backlogCmd.Flags().BoolVar(&backlogPickup, "pickup", false, "resume a backlog item instead of listing the backlog")
	backlogCmd.Flags().BoolVar(&backlogCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	backlogCmd.Flags().BoolVar(
		&backlogOutputOnly,
		"output-only",
		false,
		"output prompt text to stdout instead of copying it to the clipboard",
	)
	rootCmd.AddCommand(backlogCmd)
}

func runBacklog(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	if backlogPickup {
		outputOnly, _ := cmd.Flags().GetBool("output-only")
		var featRef string
		if len(args) == 1 {
			featRef = args[0]
		}
		return runBacklogPickup(projectRoot, cfg, specsDir, featRef, outputOnly)
	}

	if len(args) > 0 {
		return fmt.Errorf("feature arguments require --pickup")
	}
	if backlogCopy || backlogOutputOnly {
		return fmt.Errorf("--copy and --output-only can only be used with --pickup")
	}

	return printBacklogList(cmd.OutOrStdout(), specsDir, cfg)
}

func runBacklogPickup(
	projectRoot string,
	cfg *config.Config,
	specsDir string,
	featRef string,
	outputOnly bool,
) error {
	var (
		feat *feature.Feature
		err  error
	)

	if featRef != "" {
		feat, err = resolveBacklogFeature(specsDir, cfg, featRef)
		if err != nil {
			return err
		}
	} else {
		feat, err = selectBacklogFeature(specsDir, cfg, "Select a backlog item to pick up:")
		if err != nil {
			return err
		}
	}

	return resumeBacklogFeature(projectRoot, cfg, feat, outputOnly, backlogCopy, "", "backlog pickup")
}

func printBacklogList(w io.Writer, specsDir string, cfg *config.Config) error {
	entries, err := loadBacklogEntries(specsDir, cfg)
	if err != nil {
		return err
	}

	style := styleForWriter(w)
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.title("🗂️", "Backlog")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if len(entries) == 0 {
		if _, err := fmt.Fprintln(w, style.muted("No backlog items. Run `kit brainstorm --backlog` to capture deferred work.")); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		return nil
	}

	if err := printBacklogTable(w, entries); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.muted("Run `kit resume <feature>` or `kit backlog --pickup <feature>` to resume one.")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	return nil
}

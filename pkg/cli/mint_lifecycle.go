package cli

import (
	"context"
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/minttool"
	"github.com/spf13/cobra"
)

var maintainInstalledMint = minttool.MaintainInstalled

func init() { rootCmd.AddCommand(newInstallCommand()) }

func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "install", Short: "Check and update supported installed local tools",
		Args: cobra.NoArgs, SilenceUsage: true,
		Long: `Maintain optional local tools used by Kit. Currently checks installed Mint
against official stable releases and updates supported user-owned Go-bin installs
when older. Missing Mint is reported and never installed by this lifecycle command.
Current/newer versions remain unchanged; unsupported installations need their owner.

Updates run automatically, including noninteractive runs; --dry-run validates the
proposed action without installing. Repository dependencies and CI pins are separate
and never changed. This does not install/upgrade Kit or initialize a project; use
Go installation, kit upgrade, or kit init for those existing responsibilities.`,
	}
	cmd.Flags().Bool("dry-run", false, "preview local tool maintenance without installing")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		preview, _ := cmd.Flags().GetBool("dry-run")
		return runMintLifecycle(cmd, preview)
	}
	return cmd
}

func runMintLifecycle(cmd *cobra.Command, preview bool) error {
	parent := cmd.Context()
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, mintCommandTimeout(!preview))
	defer cancel()
	report, err := maintainInstalledMint(ctx, minttool.Options{}, preview)
	if err != nil {
		return fmt.Errorf("mint local maintenance failed: %w; inspect `kit mint check --json` and update unsupported installs through their owning installer; repository/CI pins remain separate", err)
	}
	message := "Mint local: " + report.VersionStatus + " (" + report.Installed.Version + "); repository/CI pins unchanged."
	switch {
	case report.Installed.Status == "missing":
		message = "Mint is not installed; skipped. Install explicitly with `kit mint update` when needed; repository/CI pins unchanged."
	case report.Updated:
		message = "Updated local Mint to " + report.Installed.Version + "; repository/CI pins unchanged."
	case preview && report.UpdateAvailable:
		message = "Preview: would update local Mint from " + report.Installed.Version + " to " + report.Target + "; no installation or repository/CI pin changes."
	}
	_, err = fmt.Fprintln(cmd.ErrOrStderr(), message)
	return err
}

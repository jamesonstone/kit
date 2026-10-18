package cli

import (
	"context"
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/minttool"
	"github.com/spf13/cobra"
)

var maintainInstalledMint = minttool.MaintainInstalled

func runMintLifecycle(cmd *cobra.Command, preview bool) error {
	parent := cmd.Context()
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, mintCommandTimeout(!preview))
	defer cancel()
	report, err := maintainInstalledMint(ctx, preview)
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

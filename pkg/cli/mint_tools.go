package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/minttool"
	"github.com/spf13/cobra"
)

var checkMintTool = minttool.Check
var updateMintTool = minttool.Update
var previewMintTool = minttool.PreviewUpdate

type mintToolsReport struct {
	Error                 string              `json:"error,omitempty"`
	Local                 minttool.Report     `json:"local"`
	RepositoryRoot        string              `json:"repository_root,omitempty"`
	RepositoryPins        []mintRepositoryPin `json:"repository_pins"`
	RepositoryPinsUpdated bool                `json:"repository_pins_updated"`
	Note                  string              `json:"note"`
}

func init() {
	group := &cobra.Command{Use: "mint", Short: "Inspect Mint versions and explicitly update a supported local installation", Args: cobra.NoArgs}
	group.AddCommand(newMintToolsCommand(false), newMintToolsCommand(true))
	rootCmd.AddCommand(group)
}

func newMintToolsCommand(update bool) *cobra.Command {
	name, short := "check", "Check the resolved Mint executable, release version, and repository pins"
	if update {
		name, short = "update", "Explicitly install a published Mint version into its verified Go bin"
	}
	cmd := &cobra.Command{Use: name, Short: short, Args: cobra.NoArgs, SilenceUsage: true,
		Long: `Inspect the locally resolved Mint executable and official published release metadata.
Repository/CI action refs and the Go module dependency are reported separately;
this command never rewrites them or implies a local update changes CI.

Check is read-only. Update explicitly installs through Mint's documented official
Go module mechanism into a verified user-owned Go bin. The candidate is staged
and verified before replacement. Unknown/package-manager/symlink installations,
development versions, downgrades, unsupported platforms and unsafe ownership fail.
No sudo or automatic privilege escalation is used. --dry-run never installs.

--require-version checks a numeric stable minimum, not feature support.
--require-linear fails until Kit records a verified published Mint release that
contains Linear support. An unmerged PR or an arbitrary newer tag is not proof.
Repository SHA pins need independent release mapping and reviewed upgrades.
Use --json for one report even on operational failure; errors exit nonzero.`,
	}
	cmd.Flags().String("version", "", "published stable vX.Y.Z target; omitted means latest stable release")
	cmd.Flags().String("require-version", "", "require at least this stable vX.Y.Z version")
	cmd.Flags().Bool("require-linear", false, "require a verified released local Mint version with Linear link support")
	cmd.Flags().Bool("json", false, "emit a structured local report and separate repository pin inventory")
	if update {
		cmd.Flags().Bool("dry-run", false, "preview through read-only checks; do not install")
	}
	cmd.RunE = func(cmd *cobra.Command, _ []string) error { return runMintTools(cmd, update) }
	return cmd
}

func runMintTools(cmd *cobra.Command, update bool) error {
	version, _ := cmd.Flags().GetString("version")
	minimum, _ := cmd.Flags().GetString("require-version")
	linear, _ := cmd.Flags().GetBool("require-linear")
	opts := minttool.Options{TargetVersion: version, RequireVersion: minimum, RequireLinear: linear}
	root, found, err := config.FindProjectRootOptional()
	if err != nil {
		return writeMintToolsReport(cmd, mintToolsReport{}, err)
	}
	var pins []mintRepositoryPin
	if found {
		pins, err = inspectMintRepositoryPins(root)
		if err != nil {
			return writeMintToolsReport(cmd, mintToolsReport{RepositoryRoot: root}, err)
		}
	} else {
		root = ""
	}
	dry := false
	if update {
		dry, _ = cmd.Flags().GetBool("dry-run")
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), mintCommandTimeout(update && !dry))
	defer cancel()
	var result minttool.Report
	if update && !dry {
		result, err = updateMintTool(ctx, opts)
	} else if update {
		result, err = previewMintTool(ctx, opts)
	} else {
		result, err = checkMintTool(ctx, opts)
	}
	report := mintToolsReport{Local: result, RepositoryRoot: root, RepositoryPins: pins, Note: "Local executable only; repository pins are inventory, not verified release mappings or automatic updates."}
	if err == nil && !update && (result.Installed.Status == "missing" || result.Installed.Status == "unsupported") {
		err = fmt.Errorf("mint local installation is %s", result.Installed.Status)
	}
	return writeMintToolsReport(cmd, report, err)
}

func writeMintToolsReport(cmd *cobra.Command, report mintToolsReport, operationErr error) error {
	if report.RepositoryPins == nil {
		report.RepositoryPins = []mintRepositoryPin{}
	}
	result, pins := report.Local, report.RepositoryPins
	if operationErr != nil {
		report.Error = operationErr.Error()
	}
	structured, _ := cmd.Flags().GetBool("json")
	if structured {
		if writeErr := json.NewEncoder(cmd.OutOrStdout()).Encode(report); writeErr != nil {
			return writeErr
		}
	} else {
		if _, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Mint local: %s %s (%s)\nLatest published: %s; target: %s; version status: %s; update available: %t; compatibility: %s; updated: %t\n", result.Installed.Path, result.Installed.Version, result.Installed.Status, result.Latest, result.Target, result.VersionStatus, result.UpdateAvailable, result.Compatibility, result.Updated); writeErr != nil {
			return writeErr
		}
		for _, pin := range pins {
			if _, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Repository %s: %s @ %s (reported only)\n", pin.Kind, pin.File, pin.Ref); writeErr != nil {
				return writeErr
			}
		}
		if _, writeErr := fmt.Fprintln(cmd.OutOrStdout(), report.Note); writeErr != nil {
			return writeErr
		}
	}
	return operationErr
}

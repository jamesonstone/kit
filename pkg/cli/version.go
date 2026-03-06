package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var buildInfoReader = debug.ReadBuildInfo

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the installed Kit version",
	Long: `Print the current Kit version from the installed binary.

This command is script-friendly and uses the same linker-injected version
value that powers the existing --version flag, with Go build info as a
fallback for module-installed binaries.`,
	Args: cobra.NoArgs,
	RunE: runVersion,
}

func init() {
	rootCmd.Version = currentVersion()
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) error {
	_, err := fmt.Fprintln(cmd.OutOrStdout(), currentVersion())
	return err
}

func currentVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}

	if buildInfo, ok := buildInfoReader(); ok {
		if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
			return buildInfo.Main.Version
		}
	}

	if Version != "" {
		return Version
	}

	return "dev"
}

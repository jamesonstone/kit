package cli

import (
	"crypto/sha256"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var upgradeYes bool

func init() {
	rootCmd.AddCommand(newUpgradeCommand("upgrade", []string{"update"}))
	rootCmd.AddCommand(newUpgradeCommand("update", nil))
}

func newUpgradeCommand(use string, aliases []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          use,
		Aliases:      aliases,
		Short:        "Download and install the latest Kit release",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE:         runUpgrade,
	}
	cmd.Flags().BoolVarP(&upgradeYes, "yes", "y", false, "skip confirmation prompt")
	return cmd
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	current := currentVersion()
	release, err := latestStableRelease(current)
	if err != nil {
		return err
	}

	latest := displayVersion(release.TagName)
	if compareVersions(current, release.TagName) >= 0 {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "✓ kit is already at %s\n", latest)
		return err
	}

	if _, err := fmt.Fprintf(
		cmd.OutOrStdout(),
		"kit %s → %s available\n\n",
		displayVersion(current),
		latest,
	); err != nil {
		return err
	}

	if !upgradeYes {
		ok, err := confirmUpgrade(cmd)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	assetName, err := selectAssetName(release.TagName, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return fmt.Errorf("%w. install manually with `%s`", err, manualInstallHint)
	}

	assetURL, ok := findAssetURL(release.Assets, assetName)
	if !ok {
		return fmt.Errorf(
			"no release asset for %s/%s: expected %s. install manually with `%s`",
			runtime.GOOS,
			runtime.GOARCH,
			assetName,
			manualInstallHint,
		)
	}

	checksumsURL, ok := findAssetURL(release.Assets, "checksums.txt")
	if !ok {
		return fmt.Errorf("release %s is missing checksums.txt", latest)
	}

	archiveBytes, err := downloadBytes(assetURL, current)
	if err != nil {
		return err
	}
	checksumBytes, err := downloadBytes(checksumsURL, current)
	if err != nil {
		return err
	}

	checksums, err := parseChecksums(string(checksumBytes))
	if err != nil {
		return err
	}
	expectedHash, ok := checksums[assetName]
	if !ok {
		return fmt.Errorf("checksums.txt is missing %s", assetName)
	}
	actualHash := fmt.Sprintf("%x", sha256.Sum256(archiveBytes))
	if expectedHash != actualHash {
		return fmt.Errorf(
			"checksum mismatch for %s: expected %s, got %s",
			assetName,
			expectedHash,
			actualHash,
		)
	}

	newBinary, err := extractBinary(archiveBytes, runtime.GOOS)
	if err != nil {
		return err
	}

	execPath, err := buildExecutablePath()
	if err != nil {
		return err
	}
	if err := replaceExecutable(execPath, newBinary); err != nil {
		return err
	}

	_, err = fmt.Fprintf(
		cmd.OutOrStdout(),
		"✅ Updated kit %s → %s\n",
		displayVersion(current),
		latest,
	)
	return err
}

package cli

func utilityCapabilityRecords() []capabilityRecord {
	return []capabilityRecord{
		capability("upgrade", "Utilities", "Upgrade the Kit CLI installation.", mutationNetwork, withNetwork("queries GitHub release metadata and downloads the selected release asset plus checksums"), withFileWrites("replaces the current Kit executable after checksum verification"), withFlags(flag("--yes", "skip the installation confirmation prompt", "installation still verifies the downloaded checksum")), withRelated(related("version", "shows current installed version")), withWhenToUse("Use when the installed Kit binary should be upgraded to the latest stable release."), withWhenNotToUse("Do not use for a read-only update check; the command installs when a newer compatible release is available and confirmation succeeds."), withExamples("kit upgrade", "kit upgrade --yes"), withCaveats("If the installed version is already current, no executable is replaced.")),
		capability("version", "Utilities", "Print the Kit CLI version.", mutationNone, withRelated(related("upgrade", "updates the installed version"))),
		capability("completion", "Utilities", "Generate shell completion scripts.", mutationNone, withFileWrites("none by default", "the shell may redirect output to a completion file outside Kit"), withRelated(related("help", "shows command syntax"))),
		capability("help", "Utilities", "Show command help and flag syntax.", mutationNone, withRelated(related("capabilities", "adds behavior and safety metadata"))),
	}
}

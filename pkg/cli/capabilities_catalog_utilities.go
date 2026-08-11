package cli

func utilityCapabilityRecords() []capabilityRecord {
	return []capabilityRecord{
		capability("upgrade", "Utilities", "Upgrade the Kit CLI installation.", mutationNetwork, withNetwork("downloads release metadata or binaries"), withFileWrites("writes the installed Kit binary or related install files"), withFlags(flag("--check", "check for an upgrade without installing when supported", "prefer for read-only inspection")), withRelated(related("version", "shows current installed version"))),
		capability("version", "Utilities", "Print the Kit CLI version.", mutationNone, withRelated(related("upgrade", "updates the installed version"))),
		capability("completion", "Utilities", "Generate shell completion scripts.", mutationNone, withFileWrites("none by default", "the shell may redirect output to a completion file outside Kit"), withRelated(related("help", "shows command syntax"))),
		capability("help", "Utilities", "Show command help and flag syntax.", mutationNone, withRelated(related("capabilities", "adds behavior and safety metadata"))),
	}
}

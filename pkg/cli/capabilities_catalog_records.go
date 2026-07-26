package cli

func capabilityCatalogRecords() []capabilityRecord {
	return append(capabilityCatalogRecordsPart1(), capabilityCatalogRecordsPart2()...)
}

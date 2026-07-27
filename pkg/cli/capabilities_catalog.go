package cli

import (
	"sort"
)

const capabilitiesSchemaVersion = 1

func capabilityCatalog() []capabilityRecord {
	records := capabilityCatalogRecords()
	sort.SliceStable(records, func(i, j int) bool {
		return lessCapabilityRecord(records[i], records[j])
	})
	return records
}

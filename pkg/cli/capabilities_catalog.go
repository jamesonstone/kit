package cli

import "sort"

const capabilitiesSchemaVersion = 1

func capabilityCatalog() []capabilityRecord {
	records := setupCapabilityRecords()
	records = append(records, workflowCapabilityRecords()...)
	records = append(records, inspectionCapabilityRecords()...)
	records = append(records, utilityCapabilityRecords()...)
	sort.SliceStable(records, func(i, j int) bool {
		return lessCapabilityRecord(records[i], records[j])
	})
	return records
}

package cli

import "strings"

func compactCapabilityRecords(records []capabilityRecord) []capabilityCompactRecord {
	compact := make([]capabilityCompactRecord, 0, len(records))
	for _, record := range records {
		if !record.IncludeInCompact || record.Hidden || record.Deprecated {
			continue
		}
		compact = append(compact, record.compact())
	}
	return compact
}

func detailCapabilityRecords(records []capabilityRecord) []capabilityDetailRecord {
	detail := make([]capabilityDetailRecord, 0, len(records))
	for _, record := range records {
		detail = append(detail, record.detail())
	}
	return detail
}

func capabilityByCommandPath(commandPath string) (capabilityRecord, bool) {
	normalized := normalizeCapabilityQuery(commandPath)
	for _, record := range capabilityCatalog() {
		if record.matchesCommandPath(normalized) {
			return record, true
		}
	}
	return capabilityRecord{}, false
}

func searchCapabilityRecords(query string) []capabilityRecord {
	normalized := normalizeCapabilityQuery(query)
	if normalized == "" {
		return visibleCapabilityRecords()
	}

	var matches []capabilityRecord
	for _, record := range visibleCapabilityRecords() {
		if record.matchesSearch(normalized) {
			matches = append(matches, record)
		}
	}
	return matches
}

func visibleCapabilityRecords() []capabilityRecord {
	records := capabilityCatalog()
	visible := make([]capabilityRecord, 0, len(records))
	for _, record := range records {
		if record.IncludeInCompact && !record.Hidden && !record.Deprecated {
			visible = append(visible, record)
		}
	}
	return visible
}

func suggestCapabilityCommands(commandPath string) []string {
	normalized := normalizeCapabilityQuery(commandPath)
	if normalized == "" {
		return nil
	}

	suggestions := make([]string, 0, 3)
	for _, record := range capabilityCatalog() {
		candidates := append([]string{record.Command}, record.Aliases...)
		for _, candidate := range candidates {
			candidate = normalizeCapabilityQuery(candidate)
			if strings.HasPrefix(candidate, normalized) || strings.HasPrefix(normalized, candidate) || strings.Contains(candidate, normalized) {
				suggestions = appendUniqueSuggestion(suggestions, record.Command)
			}
			if len(suggestions) >= 3 {
				return suggestions
			}
		}
	}
	return suggestions
}

func appendUniqueSuggestion(suggestions []string, command string) []string {
	for _, existing := range suggestions {
		if existing == command {
			return suggestions
		}
	}
	return append(suggestions, command)
}

func (record capabilityRecord) compact() capabilityCompactRecord {
	return capabilityCompactRecord{
		Command:         record.Command,
		Category:        record.Category,
		Summary:         record.Summary,
		MutationLevel:   record.MutationLevel,
		NetworkUse:      record.NetworkUse,
		FileWrites:      record.FileWrites,
		GitMutation:     record.GitMutation,
		Hidden:          record.Hidden,
		Deprecated:      record.Deprecated,
		ImportantFlags:  cloneCapabilityFlags(record.ImportantFlags),
		RelatedCommands: cloneRelatedCommands(record.RelatedCommands),
	}
}

func (record capabilityRecord) detail() capabilityDetailRecord {
	return capabilityDetailRecord{
		Command:              record.Command,
		Category:             record.Category,
		Summary:              record.Summary,
		MutationLevel:        record.MutationLevel,
		NetworkUse:           record.NetworkUse,
		FileWrites:           record.FileWrites,
		GitMutation:          record.GitMutation,
		Hidden:               record.Hidden,
		Deprecated:           record.Deprecated,
		DeprecationNote:      record.DeprecationNote,
		Aliases:              cloneStrings(record.Aliases),
		ImportantFlags:       cloneCapabilityFlags(record.ImportantFlags),
		RelatedCommands:      cloneRelatedCommands(record.RelatedCommands),
		WhenToUse:            cloneStrings(record.WhenToUse),
		WhenNotToUse:         cloneStrings(record.WhenNotToUse),
		Examples:             cloneStrings(record.Examples),
		Caveats:              cloneStrings(record.Caveats),
		DetailedFlagBehavior: cloneCapabilityFlags(record.DetailedFlagBehavior),
	}
}

func lessCapabilityRecord(left, right capabilityRecord) bool {
	leftCategory := capabilityCategoryOrder[left.Category]
	rightCategory := capabilityCategoryOrder[right.Category]
	if leftCategory != rightCategory {
		return leftCategory < rightCategory
	}

	leftOrder := capabilityCommandOrder(left.Command)
	rightOrder := capabilityCommandOrder(right.Command)
	if leftOrder != rightOrder {
		return leftOrder < rightOrder
	}
	return left.Command < right.Command
}

func capabilityCommandOrder(commandPath string) int {
	rootName := commandPath
	if fields := strings.Fields(commandPath); len(fields) > 0 {
		rootName = fields[0]
	}
	if order, ok := commandOrder[rootName]; ok {
		return order
	}
	return 1000
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func cloneCapabilityFlags(values []capabilityFlag) []capabilityFlag {
	if len(values) == 0 {
		return []capabilityFlag{}
	}
	return append([]capabilityFlag(nil), values...)
}

func cloneRelatedCommands(values []capabilityRelatedCommand) []capabilityRelatedCommand {
	if len(values) == 0 {
		return []capabilityRelatedCommand{}
	}
	return append([]capabilityRelatedCommand(nil), values...)
}

func (record capabilityRecord) matchesCommandPath(commandPath string) bool {
	if normalizeCapabilityQuery(record.Command) == commandPath {
		return true
	}
	for _, alias := range record.Aliases {
		if normalizeCapabilityQuery(alias) == commandPath {
			return true
		}
	}
	return false
}

func (record capabilityRecord) matchesSearch(query string) bool {
	searchable := []string{
		record.Command,
		record.Category,
		record.Summary,
		record.MutationLevel,
		record.NetworkUse.Summary,
		record.NetworkUse.FlagDependent,
		record.FileWrites.Summary,
		record.FileWrites.FlagDependent,
		record.GitMutation.Summary,
		record.GitMutation.FlagDependent,
	}
	searchable = append(searchable, record.Aliases...)
	searchable = append(searchable, record.WhenToUse...)
	searchable = append(searchable, record.WhenNotToUse...)
	searchable = append(searchable, record.Examples...)
	searchable = append(searchable, record.Caveats...)
	for _, flag := range record.ImportantFlags {
		searchable = append(searchable, flag.Name, flag.Summary, flag.Safety)
	}
	for _, related := range record.RelatedCommands {
		searchable = append(searchable, related.Command, related.Note)
	}

	for _, value := range searchable {
		if strings.Contains(normalizeCapabilityQuery(value), query) {
			return true
		}
	}
	return false
}

func normalizeCapabilityQuery(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

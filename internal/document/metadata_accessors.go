package document

import (
	"fmt"
	"strings"
)

func (d *Document) Relationships() ([]Relationship, []RelationshipParseWarning) {
	if d.Metadata != nil && len(d.Metadata.Relationships) > 0 {
		return relationshipsFromMetadata(d.Metadata.Relationships), nil
	}
	if d.FrontMatterPresent && !hasLegacyRelationshipSection(d.GetSection("RELATIONSHIPS")) {
		return nil, nil
	}
	return ParseRelationshipsSectionRelaxed(d.GetSection("RELATIONSHIPS"))
}

func (d *Document) References() []MetadataReference {
	if d.Metadata != nil && len(d.Metadata.References) > 0 {
		return append([]MetadataReference{}, d.Metadata.References...)
	}
	return nil
}

func (d *Document) Skills() []MetadataSkill {
	if d.Metadata != nil && len(d.Metadata.Skills) > 0 {
		return append([]MetadataSkill{}, d.Metadata.Skills...)
	}
	return SkillsFromSection(d.GetSection("SKILLS"))
}

func (d *Document) SummaryText() string {
	if d.Metadata != nil && strings.TrimSpace(d.Metadata.Summary) != "" {
		return strings.TrimSpace(d.Metadata.Summary)
	}
	if section := d.GetSection("SUMMARY"); section != nil {
		return ExtractFirstParagraph(section)
	}
	return ""
}

func (d *Document) IntentText(sectionName string) string {
	if d.Metadata != nil && strings.TrimSpace(d.Metadata.Intent) != "" {
		return strings.TrimSpace(d.Metadata.Intent)
	}
	if section := d.GetSection(sectionName); section != nil {
		return ExtractFirstParagraph(section)
	}
	return ""
}

func (d *Document) DeliveryIntent() string {
	if d.Metadata != nil {
		return strings.TrimSpace(d.Metadata.DeliveryIntent)
	}
	return ""
}

func (d *Document) ClarificationState() (MetadataClarification, bool) {
	if d.Metadata == nil || d.Metadata.Clarification == nil {
		return MetadataClarification{}, false
	}
	clarification := MetadataClarification{
		Status: strings.TrimSpace(d.Metadata.Clarification.Status),
	}
	if value, ok := d.Metadata.Clarification.ConfidenceValue(); ok {
		clarification.Confidence = intPtr(value)
	}
	if value, ok := d.Metadata.Clarification.UnresolvedQuestionsValue(); ok {
		clarification.UnresolvedQuestions = intPtr(value)
	}
	return clarification, true
}

func SkillsFromSection(section *Section) []MetadataSkill {
	if section == nil {
		return nil
	}

	rows := metadataTableRows(section.Content)
	if len(rows) < 3 {
		return nil
	}

	header := metadataHeaderIndex(rows[0])
	required := []string{"skill", "source", "path", "trigger", "required"}
	for _, key := range required {
		if _, ok := header[key]; !ok {
			return nil
		}
	}

	var skills []MetadataSkill
	for _, row := range rows[2:] {
		name := metadataCell(row, header["skill"])
		if name == "" || strings.EqualFold(normalizeMetadataCell(name), "none") {
			continue
		}
		skills = append(skills, MetadataSkill{
			Name:    name,
			Source:  metadataCell(row, header["source"]),
			Path:    metadataCell(row, header["path"]),
			Trigger: metadataCell(row, header["trigger"]),
			Required: strings.EqualFold(normalizeMetadataCell(metadataCell(row, header["required"])), "yes") ||
				strings.EqualFold(normalizeMetadataCell(metadataCell(row, header["required"])), "true"),
		})
	}
	return skills
}

func relationshipsFromMetadata(entries []MetadataRelationship) []Relationship {
	relationships := make([]Relationship, 0, len(entries))
	for _, entry := range entries {
		human, ok := RelationshipMachineToHuman(entry.Type)
		if !ok || strings.TrimSpace(entry.Target) == "" {
			continue
		}
		relationships = append(relationships, Relationship{
			Type:   human,
			Target: strings.TrimSpace(entry.Target),
		})
	}
	return relationships
}

func hasLegacyRelationshipSection(section *Section) bool {
	if section == nil {
		return false
	}
	content := strings.TrimSpace(section.Content)
	if content == "" || strings.EqualFold(content, "none") {
		return true
	}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			return true
		}
	}
	return false
}

func (d *Document) relationshipConflicts() []MetadataConflict {
	if d.Metadata == nil || len(d.Metadata.Relationships) == 0 {
		return nil
	}
	bodyRelationships, warnings := ParseRelationshipsSectionRelaxed(d.GetSection("RELATIONSHIPS"))
	if len(warnings) > 0 || len(bodyRelationships) == 0 {
		return nil
	}
	frontRelationships := relationshipsFromMetadata(d.Metadata.Relationships)
	if relationshipSetKey(frontRelationships) == relationshipSetKey(bodyRelationships) {
		return nil
	}
	return []MetadataConflict{{
		Field:      "relationships",
		FrontValue: relationshipSetKey(frontRelationships),
		BodyValue:  relationshipSetKey(bodyRelationships),
		Message:    "front matter relationships differ from legacy body relationships",
	}}
}

func (d *Document) skillConflicts() []MetadataConflict {
	if d.Metadata == nil || len(d.Metadata.Skills) == 0 {
		return nil
	}
	bodySkills := SkillsFromSection(d.GetSection("SKILLS"))
	if len(bodySkills) == 0 {
		return nil
	}
	if skillSetKey(d.Metadata.Skills) == skillSetKey(bodySkills) {
		return nil
	}
	return []MetadataConflict{{
		Field:      "skills",
		FrontValue: skillSetKey(d.Metadata.Skills),
		BodyValue:  skillSetKey(bodySkills),
		Message:    "front matter skills differ from legacy body skills",
	}}
}

func metadataTableRows(content string) [][]string {
	var rows [][]string
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(line, "|") || !strings.Contains(strings.Trim(line, "|"), "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		rows = append(rows, cells)
	}
	return rows
}

func metadataHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, cell := range header {
		index[strings.ToLower(strings.TrimSpace(cell))] = i
	}
	return index
}

func metadataCell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func normalizeMetadataCell(value string) string {
	trimmed := strings.TrimSpace(value)
	for strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") && len(trimmed) >= 2 {
		trimmed = strings.TrimSpace(strings.Trim(trimmed, "`"))
	}
	return strings.ToLower(trimmed)
}

func relationshipSetKey(relationships []Relationship) string {
	parts := make([]string, 0, len(relationships))
	for _, relationship := range relationships {
		machine, ok := RelationshipHumanToMachine(relationship.Type)
		if !ok {
			machine = normalizeMetadataCell(relationship.Type)
		}
		parts = append(parts, fmt.Sprintf("%s:%s", machine, normalizeMetadataCell(relationship.Target)))
	}
	return sortedKey(parts)
}

func referenceSetKey(references []MetadataReference) string {
	parts := make([]string, 0, len(references))
	for _, reference := range references {
		parts = append(parts, fmt.Sprintf(
			"%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
			normalizeMetadataCell(reference.ID),
			normalizeMetadataCell(reference.Name),
			normalizeMetadataCell(reference.Type),
			normalizeMetadataCell(reference.Target),
			normalizeMetadataCell(reference.SelectorType),
			normalizeMetadataCell(reference.Selector),
			normalizeMetadataCell(reference.Relation),
			normalizeMetadataCell(reference.ReadPolicy),
			normalizeMetadataCell(reference.UsedFor),
			normalizeMetadataCell(reference.Status),
		))
	}
	return sortedKey(parts)
}

func skillSetKey(skills []MetadataSkill) string {
	parts := make([]string, 0, len(skills))
	for _, skill := range skills {
		parts = append(parts, fmt.Sprintf(
			"%s:%s:%s:%s:%t",
			normalizeMetadataCell(skill.Name),
			normalizeMetadataCell(skill.Source),
			normalizeMetadataCell(skill.Path),
			normalizeMetadataCell(skill.Trigger),
			skill.Required,
		))
	}
	return sortedKey(parts)
}

func sortedKey(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	for i := 1; i < len(parts); i++ {
		for j := i; j > 0 && parts[j] < parts[j-1]; j-- {
			parts[j], parts[j-1] = parts[j-1], parts[j]
		}
	}
	return strings.Join(parts, "|")
}

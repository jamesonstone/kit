package document

import (
	"fmt"
	"regexp"
	"strings"
)

type Relationship struct {
	Type   string
	Target string
}

type RelationshipParseWarning struct {
	Line    string
	Message string
}

var (
	relationshipPattern       = regexp.MustCompile(`^-\s*(builds on|depends on|related to):\s*(.+?)\s*$`)
	relationshipTargetPattern = regexp.MustCompile(`^[0-9]+-[a-z0-9][a-z0-9-]*$`)
)

func ParseRelationshipsSection(section *Section) ([]Relationship, error) {
	if section == nil {
		return nil, nil
	}

	relationships, warnings := parseRelationshipsSection(section, false)
	if len(warnings) > 0 {
		return nil, fmt.Errorf("%s", warnings[0].Message)
	}

	return relationships, nil
}

func ParseRelationshipsSectionRelaxed(section *Section) ([]Relationship, []RelationshipParseWarning) {
	if section == nil {
		return nil, nil
	}

	return parseRelationshipsSection(section, true)
}

func parseRelationshipsSection(section *Section, relaxed bool) ([]Relationship, []RelationshipParseWarning) {
	var relationships []Relationship
	var warnings []RelationshipParseWarning
	sawNone := false

	for _, rawLine := range strings.Split(section.Content, "\n") {
		line := visibleLineContent(strings.TrimSpace(rawLine))
		if line == "" {
			continue
		}

		normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(line)), ".")
		if normalized == "none" {
			if sawNone || len(relationships) > 0 {
				warning := RelationshipParseWarning{
					Line:    line,
					Message: "`none` cannot be combined with other relationship entries",
				}
				if !relaxed {
					return nil, []RelationshipParseWarning{warning}
				}
				warnings = append(warnings, warning)
				continue
			}
			sawNone = true
			continue
		}
		if sawNone {
			warning := RelationshipParseWarning{
				Line:    line,
				Message: "relationship entries cannot follow `none`",
			}
			if !relaxed {
				return nil, []RelationshipParseWarning{warning}
			}
			warnings = append(warnings, warning)
			continue
		}

		relationship, err := parseRelationshipLine(line)
		if err != nil {
			warning := RelationshipParseWarning{
				Line:    line,
				Message: err.Error(),
			}
			if !relaxed {
				return nil, []RelationshipParseWarning{warning}
			}
			warnings = append(warnings, warning)
			continue
		}

		relationships = append(relationships, relationship)
	}

	return relationships, warnings
}

func parseRelationshipLine(line string) (Relationship, error) {
	matches := relationshipPattern.FindStringSubmatch(strings.ToLower(strings.TrimSpace(line)))
	if matches == nil {
		return Relationship{}, fmt.Errorf("expected `- builds on: <feature>`, `- depends on: <feature>`, or `- related to: <feature>`, got %q", line)
	}

	target := strings.TrimSpace(matches[2])
	target = strings.Trim(target, "`")
	target = strings.TrimSpace(strings.TrimSuffix(target, "."))
	if !relationshipTargetPattern.MatchString(target) {
		return Relationship{}, fmt.Errorf("expected `- builds on: <feature>`, `- depends on: <feature>`, or `- related to: <feature>`, got %q", line)
	}

	return Relationship{
		Type:   matches[1],
		Target: target,
	}, nil
}

func requiresRelationshipSectionValidation(docType DocumentType, sectionName string) bool {
	if strings.ToUpper(sectionName) != "RELATIONSHIPS" {
		return false
	}

	switch docType {
	case TypeBrainstorm, TypeSpec:
		return true
	default:
		return false
	}
}

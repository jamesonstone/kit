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

var relationshipPattern = regexp.MustCompile(`^-\s*(builds on|depends on|related to):\s*([0-9]+-[a-z0-9][a-z0-9-]*)$`)

func ParseRelationshipsSection(section *Section) ([]Relationship, error) {
	if section == nil {
		return nil, nil
	}

	var relationships []Relationship
	sawNone := false

	for _, rawLine := range strings.Split(section.Content, "\n") {
		line := visibleLineContent(strings.TrimSpace(rawLine))
		if line == "" {
			continue
		}

		normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(line)), ".")
		if normalized == "none" {
			if sawNone || len(relationships) > 0 {
				return nil, fmt.Errorf("`none` cannot be combined with other relationship entries")
			}
			sawNone = true
			continue
		}
		if sawNone {
			return nil, fmt.Errorf("relationship entries cannot follow `none`")
		}

		matches := relationshipPattern.FindStringSubmatch(strings.ToLower(line))
		if matches == nil {
			return nil, fmt.Errorf("expected `- builds on: <feature>`, `- depends on: <feature>`, or `- related to: <feature>`, got %q", line)
		}

		relationships = append(relationships, Relationship{
			Type:   matches[1],
			Target: matches[2],
		})
	}

	return relationships, nil
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
